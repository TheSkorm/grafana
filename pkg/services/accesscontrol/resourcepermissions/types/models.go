package types

import (
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/services/accesscontrol"
)

type SetResourcePermissionCommand struct {
	Actions           []string
	Resource          string
	ResourceID        string
	ResourceAttribute string
	Permission        string
}

type SetResourcePermissionsCommand struct {
	Binding accesscontrol.Binding
	SetResourcePermissionCommand
	Hook ResourceHookFunc
}

type GetResourcePermissionsQuery struct {
	Actions           []string
	Resource          string
	ResourceID        string
	ResourceAttribute string
	OnlyManaged       bool
	InheritedScopes   []string
	User              *models.SignedInUser
}
