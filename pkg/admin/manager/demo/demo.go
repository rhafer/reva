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

package demo

import (
	"context"
	"fmt"

	userpb "github.com/cs3org/go-cs3apis/cs3/identity/user/v1beta1"
	"github.com/cs3org/reva/pkg/admin"
	"github.com/cs3org/reva/pkg/admin/manager/registry"
	"github.com/cs3org/reva/pkg/appctx"
	"github.com/cs3org/reva/pkg/errtypes"
)

func init() {
	registry.Register("demo", New)
}

type manager struct {
	catalog map[string]*userpb.User
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
	return nil
}

func (m *manager) CreateUser(ctx context.Context, user *userpb.User) (*userpb.User, error) {
	log := appctx.GetLogger(ctx)
	if user != nil && user.Username != "" {
		if user.Username == "exists" {
			log.Debug().Str("driver", "demo").Msg(fmt.Sprintf("User already exists '%s'", user.Username))
			return nil, errtypes.AlreadyExists(user.Username)
		}
		log.Debug().Str("driver", "demo").Msg(fmt.Sprintf("creating user: '%s'", user.Username))
		return user, nil
	}
	return nil, errtypes.BadRequest("username required")
}

func (m *manager) DeleteUser(ctx context.Context, uid *userpb.UserId) error {
	return errtypes.NotFound(uid.OpaqueId)
}
