// Copyright 2020 CERN
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

package conversions

import (
	"testing"

	provider "github.com/cs3org/go-cs3apis/cs3/storage/provider/v1beta1"
)

func TestNewPermissions(t *testing.T) {
	for val := int(PermissionMinInput); val <= int(PermissionMaxInput); val++ {
		_, err := NewPermissions(val)
		if err != nil {
			t.Errorf("value %d should be a valid permissions", val)
		}
	}
}

func TestNewPermissionsWithInvalidValueShouldFail(t *testing.T) {
	vals := []int{
		-1,
		int(PermissionMaxInput) + 1,
	}
	for _, v := range vals {
		_, err := NewPermissions(v)
		if err == nil {
			t.Errorf("value %d should not be a valid permission", v)
		}
	}
}

func TestContainPermissionAll(t *testing.T) {
	table := map[int]Permissions{
		1:  PermissionRead,
		2:  PermissionWrite,
		4:  PermissionCreate,
		8:  PermissionDelete,
		16: PermissionShare,
		31: PermissionAll,
	}

	p, _ := NewPermissions(31) // all permissions should contain all other permissions
	for _, value := range table {
		if !p.Contain(value) {
			t.Errorf("permissions %d should contain %d", p, value)
		}
	}
}
func TestContainPermissionRead(t *testing.T) {
	table := map[int]Permissions{
		2:  PermissionWrite,
		4:  PermissionCreate,
		8:  PermissionDelete,
		16: PermissionShare,
		31: PermissionAll,
	}

	p, _ := NewPermissions(1) // read permission should not contain any other permissions
	if !p.Contain(PermissionRead) {
		t.Errorf("permissions %d should contain %d", p, PermissionRead)
	}
	for _, value := range table {
		if p.Contain(value) {
			t.Errorf("permissions %d should not contain %d", p, value)
		}
	}
}

func TestContainPermissionCustom(t *testing.T) {
	table := map[int]Permissions{
		2:  PermissionWrite,
		8:  PermissionDelete,
		31: PermissionAll,
	}

	p, _ := NewPermissions(21) // read, create & share permission
	if !p.Contain(PermissionRead) {
		t.Errorf("permissions %d should contain %d", p, PermissionRead)
	}
	if !p.Contain(PermissionCreate) {
		t.Errorf("permissions %d should contain %d", p, PermissionCreate)
	}
	if !p.Contain(PermissionShare) {
		t.Errorf("permissions %d should contain %d", p, PermissionShare)
	}
	for _, value := range table {
		if p.Contain(value) {
			t.Errorf("permissions %d should not contain %d", p, value)
		}
	}
}

func TestContainWithMultiplePermissions(t *testing.T) {
	table := map[int][]Permissions{
		3: {
			PermissionRead,
			PermissionWrite,
		},
		5: {
			PermissionRead,
			PermissionCreate,
		},
		31: {
			PermissionRead,
			PermissionWrite,
			PermissionCreate,
			PermissionDelete,
			PermissionShare,
		},
	}

	for key, value := range table {
		p, _ := NewPermissions(key)
		for _, v := range value {
			if !p.Contain(v) {
				t.Errorf("permissions %d should contain %d", p, v)
			}
		}
	}
}

func TestPermissions2Role(t *testing.T) {
	checkRole := func(expected, actual string) {
		if actual != expected {
			t.Errorf("Expected role %s actually got %s", expected, actual)
		}
	}

	resourceInfoSpaceRoot := &provider.ResourceInfo{
		Type: provider.ResourceType_RESOURCE_TYPE_CONTAINER,
		Id: &provider.ResourceId{
			StorageId: "storageid",
			SpaceId:   "spaceid",
			OpaqueId:  "spaceid",
		},
		Space: &provider.StorageSpace{
			Root: &provider.ResourceId{
				StorageId: "storageid",
				SpaceId:   "spaceid",
				OpaqueId:  "spaceid",
			},
		},
	}
	resourceInfoDir := &provider.ResourceInfo{
		Type: provider.ResourceType_RESOURCE_TYPE_CONTAINER,
		Id: &provider.ResourceId{
			StorageId: "storageid",
			SpaceId:   "spaceid",
			OpaqueId:  "fileid",
		},
		Space: &provider.StorageSpace{
			Root: &provider.ResourceId{
				StorageId: "storageid",
				SpaceId:   "spaceid",
				OpaqueId:  "spaceid",
			},
		},
	}

	type permissionOnResourceInfo2Role struct {
		permissions  Permissions
		resourceInfo *provider.ResourceInfo
		role         string
	}

	table := []permissionOnResourceInfo2Role{
		{
			permissions:  PermissionRead,
			role:         RoleViewer,
			resourceInfo: resourceInfoDir,
		}, {
			permissions:  PermissionRead | PermissionShare,
			role:         RoleLegacy,
			resourceInfo: resourceInfoSpaceRoot,
		}, {
			permissions:  PermissionRead,
			role:         RoleSpaceViewer,
			resourceInfo: resourceInfoSpaceRoot,
		}, {
			permissions:  PermissionRead | PermissionWrite | PermissionCreate | PermissionDelete,
			role:         RoleEditor,
			resourceInfo: nil,
		}, {
			permissions:  PermissionWrite,
			role:         RoleLegacy,
			resourceInfo: nil,
		}, {
			permissions:  PermissionShare,
			role:         RoleLegacy,
			resourceInfo: nil,
		}, {
			permissions:  PermissionWrite | PermissionShare,
			role:         RoleLegacy,
			resourceInfo: nil,
		},
	}

	for _, t := range table {
		actual := RoleFromOCSPermissions(t.permissions, t.resourceInfo).Name
		checkRole(t.role, actual)
	}
}
