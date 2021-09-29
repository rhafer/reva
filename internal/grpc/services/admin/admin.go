// Copyright 2021 CERN
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

package admin

import (
	"context"
	"fmt"

	userpb "github.com/cs3org/go-cs3apis/cs3/admin/user/v1beta1"
	"github.com/cs3org/reva/pkg/admin"
	"github.com/cs3org/reva/pkg/admin/manager/registry"
	"github.com/cs3org/reva/pkg/errtypes"
	"github.com/cs3org/reva/pkg/rgrpc"
	"github.com/cs3org/reva/pkg/rgrpc/status"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

func init() {
	rgrpc.Register("admin", New)
}

type config struct {
	Driver  string                            `mapstructure:"driver"`
	Drivers map[string]map[string]interface{} `mapstructure:"drivers"`
}

func (c *config) init() {
	if c.Driver == "" {
		c.Driver = "demo"
	}
}

func parseConfig(m map[string]interface{}) (*config, error) {
	c := &config{}
	if err := mapstructure.Decode(m, c); err != nil {
		err = errors.Wrap(err, "error decoding conf")
		return nil, err
	}
	c.init()
	return c, nil
}

func getDriver(c *config) (admin.Manager, error) {
	if f, ok := registry.NewFuncs[c.Driver]; ok {
		return f(c.Drivers[c.Driver])
	}
	return nil, errtypes.NotFound(fmt.Sprintf("driver %s not found for admin manager", c.Driver))
}

type service struct {
	adminmgr admin.Manager
}

// New returns a new AdminServiceServer
func New(m map[string]interface{}, ss *grpc.Server) (rgrpc.Service, error) {
	c, err := parseConfig(m)

	adminManager, err := getDriver(c)
	if err != nil {
		return nil, err
	}
	svc := &service{
		adminmgr: adminManager,
	}

	return svc, nil
}

func (s *service) Close() error {
	return nil
}

func (s *service) UnprotectedEndpoints() []string {
	return []string{}
}

func (s *service) Register(ss *grpc.Server) {
	userpb.RegisterUserAPIServer(ss, s)
}

func (s *service) CreateUser(ctx context.Context, req *userpb.CreateUserRequest) (*userpb.CreateUserResponse, error) {
	user, err := s.adminmgr.CreateUser(ctx, req.User)
	if err != nil {
		if _, ok := err.(errtypes.AlreadyExists); ok {
			res := &userpb.CreateUserResponse{
				Status: status.NewAlreadyExists(ctx, err, "User already exists"),
			}
			return res, nil
		} else {
			err = errors.Wrap(err, "adminsvc: error creating user")
			res := &userpb.CreateUserResponse{
				Status: status.NewInternal(ctx, err, "error creating user"),
			}
			return res, nil
		}
	}

	res := &userpb.CreateUserResponse{
		Status: status.NewOK(ctx),
		User:   user,
	}
	return res, nil
}

func (s *service) DeleteUser(ctx context.Context, req *userpb.DeleteUserRequest) (*userpb.DeleteUserResponse, error) {
	err := s.adminmgr.DeleteUser(ctx, req.UserId)
	if err != nil {
		err = errors.Wrap(err, "adminvc: error deleting user")
		res := &userpb.DeleteUserResponse{
			Status: status.NewInternal(ctx, err, "error deleting user"),
		}
		return res, nil

	}
	res := &userpb.DeleteUserResponse{
		Status: status.NewOK(ctx),
	}
	return res, nil
}
