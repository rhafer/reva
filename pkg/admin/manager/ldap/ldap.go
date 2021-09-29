// Copyright 2018-2021 CERN
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// In applying this license, CERN does not waive the privileges and immunities
// granted to it by virtue of its status as an Intergovernmental Organization
// or submit itself to any jurisdiction.

package ldap

import (
	"context"
	"regexp"

	userpb "github.com/cs3org/go-cs3apis/cs3/identity/user/v1beta1"
	"github.com/cs3org/reva/pkg/admin"
	"github.com/cs3org/reva/pkg/admin/manager/registry"
	"github.com/cs3org/reva/pkg/appctx"
	"github.com/cs3org/reva/pkg/errtypes"
	"github.com/cs3org/reva/pkg/utils"
	"github.com/go-ldap/ldap/v3"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

var (
	matchPwHash = regexp.MustCompile(`^\{\w+\}.+$'`)
)

func init() {
	registry.Register("ldap", New)
}

type config struct {
	utils.LDAPConn    `mapstructure:",squash"`
	UserBaseDN        string   `mapstructure:"user_base_dn"`
	UserRdnAttribute  string   `mapstructure:"user_rdn_attribute"`
	UserObjectClasses []string `mapstructure:"user_objectclasses"`
}

type manager struct {
	c *config
}

func parseConfig(m map[string]interface{}) (*config, error) {
	c := config{}
	c.UserObjectClasses = []string{"inetOrgPerson"}

	if err := mapstructure.Decode(m, &c); err != nil {
		err = errors.Wrap(err, "error decoding conf")
		return nil, err
	}
	return &c, nil
}

// New returns a new user manager.
func New(m map[string]interface{}) (admin.Manager, error) {
	mgr := &manager{}
	err := mgr.Configure(m)
	if err != nil {
		return nil, err
	}
	return mgr, err
}

func (m *manager) Configure(ml map[string]interface{}) error {
	c, _ := parseConfig(ml)
	m.c = c
	return nil
}

func (m *manager) CreateUser(ctx context.Context, user *userpb.User) (*userpb.User, error) {
	log := appctx.GetLogger(ctx)
	l, err := utils.GetLDAPConnection(&m.c.LDAPConn)
	if err != nil {
		log.Error().Err(err).Msg("Error getting LDAP connection")
		err := errors.Wrap(err, "Error getting an LDAP connection")
		return nil, err
	}
	defer l.Close()

	// try to extract password from the Opaque user data
	userpw := ""
	use_chgpw_exop := false

	if pwObj, ok := user.Opaque.Map["password"]; ok {
		userpw = string(pwObj.Value)
	}

	userdn := m.c.UserRdnAttribute + "=" + user.Username + "," + m.c.UserBaseDN
	addRequest := ldap.NewAddRequest(userdn, nil)

	addRequest.Attribute("objectclass", m.c.UserObjectClasses)
	addRequest.Attribute("sn", []string{user.Username})
	addRequest.Attribute("cn", []string{user.Username})

	// If the password string is already hashed (using OpenLDAP's "{scheme}hashedpw"
	// syntax) add it to the Add request. Otherwise assume a cleartext password
	// and try to set it after creating the user using the LDAP Change Password EXOP
	if userpw != "" && matchPwHash.MatchString(userpw) {
		addRequest.Attribute("userpassword", []string{userpw})
	} else {
		use_chgpw_exop = true
	}

	err = l.Add(addRequest)
	if err != nil {
		log.Error().Err(err).Msg("Error during LDAPAdd")
		if lerr, ok := err.(*ldap.Error); ok {
			if lerr.ResultCode == ldap.LDAPResultEntryAlreadyExists {
				return nil, errtypes.AlreadyExists(user.Username)
			}
		}
	}

	// Set password using LDAP Modifiy Password Operation (to do serverside password
	// hashing), when a non-hashed password is supplied
	if use_chgpw_exop {
		setPwRequest := ldap.NewPasswordModifyRequest(userdn, "", userpw)
		_, err = l.PasswordModify(setPwRequest)

		if err != nil {
			log.Error().Err(err).Msg("Error when setting password")
			if lerr, ok := err.(*ldap.Error); ok {
				if lerr.ResultCode == ldap.LDAPResultEntryAlreadyExists {
					return nil, errtypes.AlreadyExists(user.Username)
				}
			}
		}
	}
	return user, nil
}

func (m *manager) DeleteUser(ctx context.Context, uid *userpb.UserId) error {
	return errtypes.NotFound(uid.OpaqueId)
}
