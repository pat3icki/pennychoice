package users

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/pat3icki/encrypt"
	"github.com/stretchr/testify/require"
)

func TestCreateProfile(t *testing.T) {
	tests := createProfileTestsFunc(t)
	manager := GlobalTestManager
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Handle nil profile separately because json.Marshal(nil) returns "null" with no error
			if tt.profile == nil {
				// The manager should receive a nil payload? Probably not; we'll skip or handle.
				// In this test we expect an error, so we can simulate by passing empty data.
				data := []byte("null")
				err := manager.CreateProfile(t.Context(), &CreateUserProfileParams{
					Schema:  "create_user.schema",
					Payload: data,
				})
				if tt.expectError {
					require.Error(t, err)
				} else {
					fmt.Println(err)
					require.NoError(t, err)
				}
				return
			}

			data, err := json.Marshal(tt.profile)
			require.NoError(t, err, "json.Marshal should not fail for valid profile")

			err = manager.CreateProfile(t.Context(), &CreateUserProfileParams{
				Schema:  "create_user.schema",
				Payload: data,
			})
			if tt.expectError {
				require.Error(t, err, "expected error but got none")
			} else {
				require.NoError(t, err, "unexpected error")
			}
		})
	}
}

func TestCreateContestantProfile(t *testing.T) {
	tests := createContestantProfileTestsFunc(t)

	manager := GlobalTestManager

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.profile == nil {
				data := []byte("null")
				err := manager.CreateContestantProfile(t.Context(), uuid.Nil, &CreateContestantProfileParams{
					Schema:  "create_contestant.schema",
					Payload: data,
				})
				if tt.expectError {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
				return
			}
			user_profile := createRandomUserProfile(t)
			user_id := createUserProfileTest(t, manager, user_profile)

			data, err := json.Marshal(tt.profile)
			require.NoError(t, err, "json.Marshal should not fail for valid profile")

			ctx := t.Context()
			err = manager.CreateContestantProfile(ctx, user_id, &CreateContestantProfileParams{
				Schema:  "create_contestant.schema",
				Payload: data,
			})
			if tt.expectError {
				require.Error(t, err, "expected error but got none")
			} else {
				require.NoError(t, err, "unexpected error")
			}
		})
	}
}

func TestLogin(t *testing.T) {
	manager := GlobalTestManager
	key := []byte("12345678901234567890123456789012") // 32-byte key for Branca encryption

	// Create test user with email and phone
	userProfile := createRandomUserProfile(t)
	createUserProfileTest(t, manager, userProfile)

	t.Run("valid login with email mode", func(t *testing.T) {
		ctx := t.Context()
		req := LoginRequest{
			Mode:       "email",
			Identifier: userProfile.Email,
			Password:   []byte(userProfile.Password),
		}
		ret, err := manager.Login(ctx, key, req)
		require.NoError(t, err)
		require.NotNil(t, ret)
		require.NotEmpty(t, ret.AuthorizationID)

		// Verify token decryption
		decryptedPayload, _, err := encrypt.Decrypt(&encrypt.Branca{Password: string(key)}, []byte(ret.AuthorizationID))
		require.NoError(t, err)
		require.NotEmpty(t, decryptedPayload)
	})

	t.Run("valid login with phone mode", func(t *testing.T) {
		ctx := t.Context()
		req := LoginRequest{
			Mode:       "phone",
			Identifier: userProfile.Phone,
			Password:   []byte(userProfile.Password),
		}
		ret, err := manager.Login(ctx, key, req)
		require.NoError(t, err)
		require.NotNil(t, ret)
		require.NotEmpty(t, ret.AuthorizationID)
	})

	t.Run("valid login with result opts profile filtering", func(t *testing.T) {
		ctx := t.Context()
		req := LoginRequest{
			Mode:       "email",
			Identifier: userProfile.Email,
			Password:   []byte(userProfile.Password),
			ResultOpts: &LoginReqResultOpts{
				Schema: []string{"email", "first_name", "last_name"},
			},
		}
		ret, err := manager.Login(ctx, key, req)
		require.NoError(t, err)
		require.NotNil(t, ret)
		require.NotNil(t, ret.Profile)

		var filtered map[string]interface{}
		err = json.Unmarshal(*ret.Profile, &filtered)
		require.NoError(t, err)
		require.Equal(t, userProfile.Email, filtered["email"])
		require.Equal(t, userProfile.FirstName, filtered["first_name"])
		require.Equal(t, userProfile.LastName, filtered["last_name"])
		require.Nil(t, filtered["phone"])
	})

	t.Run("invalid password returns ErrInvalidCredentials", func(t *testing.T) {
		ctx := t.Context()
		req := LoginRequest{
			Mode:       "email",
			Identifier: userProfile.Email,
			Password:   []byte("WrongPassword123!"),
		}
		ret, err := manager.Login(ctx, key, req)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrInvalidCredentials)
		require.Nil(t, ret)
	})

	t.Run("non-existent identifier returns ErrInvalidCredentials", func(t *testing.T) {
		ctx := t.Context()
		req := LoginRequest{
			Mode:       "email",
			Identifier: "nonexistent_user_email@example.com",
			Password:   []byte(userProfile.Password),
		}
		ret, err := manager.Login(ctx, key, req)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrInvalidCredentials)
		require.Nil(t, ret)
	})

	t.Run("invalid login mode returns error", func(t *testing.T) {
		ctx := t.Context()
		req := LoginRequest{
			Mode:       "unsupported_mode",
			Identifier: userProfile.Email,
			Password:   []byte(userProfile.Password),
		}
		ret, err := manager.Login(ctx, key, req)
		require.Error(t, err)
		require.EqualError(t, err, "invalid login mode")
		require.Nil(t, ret)
	})
}

func TestDeactiviateUser(t *testing.T) {
	t.Skip("test not implemented")
}

func TestGetUserProfile(t *testing.T) {
	t.Skip("test not implemented")
}
