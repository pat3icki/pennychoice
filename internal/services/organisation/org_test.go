package organisation

import (
	"encoding/json"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/pat3icki/pennychoice/internal/db/sqlc"
	"github.com/stretchr/testify/require"
)

func TestCreate(t *testing.T) {
	ctx := t.Context()
	manager := GlobalTestManager
	var (
		user sqlc.User
	)
	createInitais(t, ctx, manager, &user)
	org, members, err := manager.Create(ctx, user.ID, CreateParams{
		Name:        gofakeit.Name(),
		Description: gofakeit.NounAbstract(),
	})
	require.NoError(t, err)
	require.NotNil(t, org)
	org_usr := members[0]

	require.Equal(t, user.Email, org_usr.Email)
	require.Equal(t, user.FirstName+" "+user.LastName, org_usr.Name)
}

func TestGet(t *testing.T) {
	ctx := t.Context()
	manager := GlobalTestManager
	var (
		user sqlc.User
	)
	createInitais(t, ctx, manager, &user)
	org, _, err := manager.Create(ctx, user.ID, CreateParams{
		Name:        gofakeit.Name(),
		Description: gofakeit.NounAbstract(),
	})
	require.NoError(t, err)

	retOrg, err := manager.Get(ctx, user.ID, org.ID)
	require.NoError(t, err)
	require.Equal(t, org.ID, retOrg.ID)
}

func TestGetMembers(t *testing.T) {
	ctx := t.Context()
	manager := GlobalTestManager
	var user sqlc.User
	createInitais(t, ctx, manager, &user)

	org, _, err := manager.Create(ctx, user.ID, CreateParams{
		Name:        gofakeit.Name(),
		Description: gofakeit.NounAbstract(),
	})
	require.NoError(t, err)

	auth := Auth{
		UserID:         user.ID,
		OrganisationID: org.ID,
	}

	members, err := manager.GetMembers(ctx, auth)
	require.NoError(t, err)
	require.Len(t, members, 1)
	require.Equal(t, user.Email, members[0].Email)

	var nonMemberUser sqlc.User
	createInitais(t, ctx, manager, &nonMemberUser)
	unauth := Auth{
		UserID:         nonMemberUser.ID,
		OrganisationID: org.ID,
	}
	_, err = manager.GetMembers(ctx, unauth)
	require.Error(t, err)
}

func TestGetPermission(t *testing.T) {
	ctx := t.Context()
	manager := GlobalTestManager
	var user sqlc.User
	createInitais(t, ctx, manager, &user)

	initialPerms := []byte(`{"can_create_campaign":"true"}`)
	org, _, err := manager.Create(ctx, user.ID, CreateParams{
		Name:        gofakeit.Name(),
		Description: gofakeit.NounAbstract(),
		Permissions: initialPerms,
	})
	require.NoError(t, err)

	perms, err := manager.GetPermission(ctx, org.ID)
	require.NoError(t, err)
	require.JSONEq(t, string(initialPerms), string(perms))
}

func TestUpdatePermission(t *testing.T) {
	ctx := t.Context()
	manager := GlobalTestManager
	var user sqlc.User
	createInitais(t, ctx, manager, &user)

	initialPerms := []byte(`{"can_create_campaign":"true"}`)
	org, _, err := manager.Create(ctx, user.ID, CreateParams{
		Name:        gofakeit.Name(),
		Description: gofakeit.NounAbstract(),
		Permissions: initialPerms,
	})
	require.NoError(t, err)

	updateArgs := map[string]string{
		"can_withdraw": "false",
		"max_admins":   "5",
	}

	updatedPerms, err := manager.UpdatePermission(ctx, org.ID, updateArgs)
	require.NoError(t, err)

	var result map[string]string
	err = json.Unmarshal(updatedPerms, &result)
	require.NoError(t, err)
	require.Equal(t, "true", result["can_create_campaign"])
	require.Equal(t, "false", result["can_withdraw"])
	require.Equal(t, "5", result["max_admins"])

	storedPerms, err := manager.GetPermission(ctx, org.ID)
	require.NoError(t, err)
	require.JSONEq(t, string(updatedPerms), string(storedPerms))
}
