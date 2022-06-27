package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/services/accesscontrol"
	"github.com/grafana/grafana/pkg/services/accesscontrol/resourcepermissions/types"
	"github.com/grafana/grafana/pkg/services/sqlstore"
)

type getUserPermissionsTestCase struct {
	desc               string
	orgID              int64
	role               string
	userPermissions    []string
	teamPermissions    []string
	builtinPermissions []string
	actions            []string
	expected           int
}

func TestAccessControlStore_GetUserPermissions(t *testing.T) {
	tests := []getUserPermissionsTestCase{
		{
			desc:               "should successfully get user, team and builtin permissions",
			orgID:              1,
			role:               "Admin",
			userPermissions:    []string{"1", "2", "10"},
			teamPermissions:    []string{"100", "2"},
			builtinPermissions: []string{"5", "6"},
			expected:           7,
		},
		{
			desc:               "Should not get admin roles",
			orgID:              1,
			role:               "Viewer",
			userPermissions:    []string{"1", "2", "10"},
			teamPermissions:    []string{"100", "2"},
			builtinPermissions: []string{"5", "6"},
			expected:           5,
		},
		{
			desc:               "Should work without org role",
			orgID:              1,
			role:               "",
			userPermissions:    []string{"1", "2", "10"},
			teamPermissions:    []string{"100", "2"},
			builtinPermissions: []string{"5", "6"},
			expected:           5,
		},
		{
			desc:               "Should filter on actions",
			orgID:              1,
			role:               "",
			userPermissions:    []string{"1", "2", "10"},
			teamPermissions:    []string{"100", "2"},
			builtinPermissions: []string{"5", "6"},
			expected:           3,
			actions:            []string{"dashboards:write"},
		},
		{
			desc:               "Should return no permission when passing empty slice of actions",
			orgID:              1,
			role:               "Viewer",
			userPermissions:    []string{"1", "2", "10"},
			teamPermissions:    []string{"100", "2"},
			builtinPermissions: []string{"5", "6"},
			expected:           0,
			actions:            []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			store, sql := setupTestEnv(t)

			user, team := createUserAndTeam(t, sql, tt.orgID)

			for _, id := range tt.userPermissions {
				_, err := store.SetResourcePermission(context.Background(), tt.orgID, accesscontrol.UserBinding(user.Id), types.SetResourcePermissionCommand{
					Actions:           []string{"dashboards:write"},
					Resource:          "dashboards",
					ResourceID:        id,
					ResourceAttribute: "id",
				}, nil)
				require.NoError(t, err)
			}

			for _, id := range tt.teamPermissions {
				_, err := store.SetResourcePermission(context.Background(), tt.orgID, accesscontrol.TeamBinding(team.Id), types.SetResourcePermissionCommand{
					Actions:           []string{"dashboards:read"},
					Resource:          "dashboards",
					ResourceID:        id,
					ResourceAttribute: "id",
				}, nil)
				require.NoError(t, err)
			}

			for _, id := range tt.builtinPermissions {
				_, err := store.SetResourcePermission(context.Background(), tt.orgID, accesscontrol.BuiltInRoleBinding("Admin"), types.SetResourcePermissionCommand{
					Actions:           []string{"dashboards:read"},
					Resource:          "dashboards",
					ResourceID:        id,
					ResourceAttribute: "id",
				}, nil)
				require.NoError(t, err)
			}

			permissions, err := store.GetUserPermissions(context.Background(), tt.orgID, &models.SignedInUser{
				UserId:  user.Id,
				OrgId:   user.OrgId,
				Teams:   []int64{team.Id},
				OrgRole: models.RoleType(tt.role),
			}, accesscontrol.GetUserPermissionsQuery{
				Actions: tt.actions,
			})

			require.NoError(t, err)
			assert.Len(t, permissions, tt.expected)
		})
	}
}

func createUserAndTeam(t *testing.T, sql *sqlstore.SQLStore, orgID int64) (*models.User, models.Team) {
	t.Helper()

	user, err := sql.CreateUser(context.Background(), models.CreateUserCommand{
		Login: "user",
		OrgId: orgID,
	})
	require.NoError(t, err)

	team, err := sql.CreateTeam("team", "", orgID)
	require.NoError(t, err)

	err = sql.AddTeamMember(user.Id, orgID, team.Id, false, models.PERMISSION_VIEW)
	require.NoError(t, err)

	return user, team
}

func setupTestEnv(t testing.TB) (*AccessControlStore, *sqlstore.SQLStore) {
	store := sqlstore.InitTestDB(t)
	return ProvideService(store), store
}
