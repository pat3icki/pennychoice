package users

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type createProfileTests struct {
	name        string
	profile     *CreateProfile
	expectError bool
}

func createProfileTestsFunc(t *testing.T) []createProfileTests {
	t.Helper()

	validDOB := time.Date(1995, 6, 15, 0, 0, 0, 0, time.UTC)

	return []createProfileTests{
		{
			name: "valid profile – all required fields",
			profile: &CreateProfile{
				FirstName:   "Amara",
				LastName:    "Okafor",
				DateOfBirth: validDOB,
				Gender:      "F",
				Country:     "Nigeria",
				Email:       gofakeit.Email(),
				Password:    gofakeit.Password(true, true, true, false, false, 7),
			},
			expectError: false,
		},
		{
			name: "valid profile – with optional fields",
			profile: &CreateProfile{
				FirstName:   gofakeit.FirstName(),
				MiddleName:  gofakeit.Name(),
				LastName:    gofakeit.LastName(),
				DateOfBirth: validDOB,
				Gender:      "M",
				Country:     "Nigeria",
				Email:       gofakeit.Email(),
				Phone:       gofakeit.Phone(),
				Password:    gofakeit.Password(true, true, true, false, false, 7),
				PIN:         "4821",
			},
			expectError: false,
		},
		{
			name: "missing first name",
			profile: &CreateProfile{
				LastName:    gofakeit.LastName(),
				DateOfBirth: validDOB,
				Gender:      "F",
				Country:     gofakeit.Country(),
				Email:       gofakeit.Email(),
				Password:    gofakeit.Password(true, true, true, false, false, 8),
			},
			expectError: true,
		},
		{
			name: "missing last name",
			profile: &CreateProfile{
				FirstName:   "Amara",
				DateOfBirth: validDOB,
				Gender:      "F",
				Country:     "Nigeria",
				Email:       gofakeit.Email(),
				Password:    gofakeit.Password(true, true, true, false, false, 8),
			},
			expectError: true,
		},
		{
			name: "missing email",
			profile: &CreateProfile{
				FirstName:   "Amara",
				LastName:    "Okafor",
				DateOfBirth: validDOB,
				Gender:      "F",
				Country:     "Nigeria",
				Password:    gofakeit.Password(true, true, true, false, false, 8),
			},
			expectError: true,
		},
		{
			name: "missing password",
			profile: &CreateProfile{
				FirstName:   "Amara",
				LastName:    "Okafor",
				DateOfBirth: validDOB,
				Gender:      "F",
				Country:     "Nigeria",
				Email:       gofakeit.Email(),
			},
			expectError: true,
		},
		{
			name: "password below minimum length",
			profile: &CreateProfile{
				FirstName:   "Tunde",
				LastName:    "Bello",
				DateOfBirth: validDOB,
				Gender:      "M",
				Country:     "Nigeria",
				Email:       gofakeit.Email(),
				Password:    gofakeit.Password(true, true, true, false, false, 2),
			},
			expectError: true,
		},
		{
			name: "invalid email format",
			profile: &CreateProfile{
				FirstName:   "Yemi",
				LastName:    "Adeyemi",
				DateOfBirth: validDOB,
				Gender:      "F",
				Country:     "Nigeria",
				Email:       "not-an-email",
				Password:    "ValidPass@99",
			},
			expectError: true,
		},
		{
			name:        "nil profile pointer",
			profile:     nil,
			expectError: true,
		},
	}
}

type createContestantProfileTests struct {
	name        string
	profile     *CreateContestantProfile
	expectError bool
}

// createRandomUserProfile returns a valid CreateProfile with random data.
func createRandomUserProfile(t *testing.T) *CreateProfile {
	t.Helper()
	gender := "M"
	if gofakeit.Bool() {
		gender = "F"
	}
	return &CreateProfile{
		FirstName:   gofakeit.FirstName(),
		LastName:    gofakeit.LastName(),
		DateOfBirth: gofakeit.DateRange(time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2005, 1, 1, 0, 0, 0, 0, time.UTC)),
		Gender:      gender,
		Country:     gofakeit.Country(),
		Email:       gofakeit.Email(),
		Phone:       gofakeit.Phone(),
		Password:    gofakeit.Password(true, true, true, false, false, 8),
	}
}

func createUserProfileTest(t *testing.T, manager *Service, usr *CreateProfile) uuid.UUID {
	data, err := json.Marshal(usr)
	require.NoError(t, err, "json.Marshal should not fail for valid profile")

	err = manager.CreateProfile(t.Context(), &CreateUserProfileParams{
		Schema:  "create_user.schema",
		Payload: data,
	})
	require.NoError(t, err, "error manager.CreateProfile")

	d, err := manager.Postgres.Query.GetUserByEmail(t.Context(), usr.Email)
	require.NoError(t, err, "d, err := manager.Postgres.Query.GetUserByEmail(t.Context(), usr.Email)")

	return d.ID
}

// createContestantProfileTestsFunc now uses "PLACEHOLDER" for user_id in valid cases.
func createContestantProfileTestsFunc(t *testing.T) []createContestantProfileTests {
	t.Helper()

	return []createContestantProfileTests{
		{
			name: "valid contestant profile – required fields only",
			profile: &CreateContestantProfile{
				ProfileName: "music",
				DisplayName: "DJ Flame",
				IsActive:    true,
			},
			expectError: false,
		},
		{
			name: "valid contestant profile – all optional fields",
			profile: &CreateContestantProfile{
				ProfileName: "art",
				DisplayName: "Canvas Queen",
				Bio:         "Contemporary visual artist from Lagos.",
				PhotoRoot:   "https://cdn.example.com/photos/canvas-queen",
				CoverPhoto:  "https://cdn.example.com/covers/canvas-queen",
				// SocialLinks:      []byte(`["https://instagram.com/canvasqueen", "https://twitter.com/canvasqueen"]`), // array, not object
				WebsiteUrl:       "https://canvasqueen.art",
				Location:         "Lagos, Nigeria",
				IsActive:         true,
				OrganisationName: gofakeit.Name(),
				Category:         "visual-art",
			},
			expectError: false,
		},
		{
			name: "missing profile_name",
			profile: &CreateContestantProfile{
				DisplayName: "Silent Thunder",
				IsActive:    true,
			},
			expectError: true,
		},
		{
			name: "missing display_name",
			profile: &CreateContestantProfile{
				ProfileName: "dance",
				IsActive:    true,
			},
			expectError: true,
		},
		{
			name:        "nil profile pointer",
			profile:     nil,
			expectError: true,
		},
	}
}
