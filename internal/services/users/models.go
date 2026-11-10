package users

import (
	"context"
	"encoding/json"
	"time"

	"github.com/pat3icki/filter"

	"github.com/pat3icki/pennychoice/gmodels"
	"github.com/pat3icki/pennychoice/internal/db/sqlc"
)

type UserServiceImpl interface {
	CreateProfile(ctx context.Context, params *CreateUserProfileParams) error
	CreateContestantProfile(ctx context.Context, params *CreateContestantProfileParams) error
	Login(ctx context.Context, key []byte, params LoginRequest) (*LoginResponse, error)
	InitDeactivation(ctx context.Context, usr gmodels.User) error
	Deactivate(ctx context.Context, usr gmodels.User, code string) error

	RequestPasswordReset()
	ResetPassword()
}

type CreateUserProfileParams struct {
	Schema  string
	Payload []byte
}

type CreateContestantProfileParams struct {
	Schema  string
	Payload []byte
}

// CreateContestantProfile is the JSON-decoded body for CreateContestantProfile.
type CreateContestantProfile struct {
	ProfileName      string          `json:"profile_name"`
	DisplayName      string          `json:"display_name"`
	Bio              string          `json:"bio,omitempty"`
	PhotoRoot        string          `json:"photo_root,omitempty"`
	CoverPhoto       string          `json:"cover_photo,omitempty"`
	SocialLinks      json.RawMessage `json:"social_links,omitempty"`
	WebsiteUrl       string          `json:"website_url,omitempty"`
	Location         string          `json:"location,omitempty"`
	IsActive         bool            `json:"is_active"`
	OrganisationName string          `json:"organisation_name,omitempty"`
	Category         string          `json:"category,omitempty"`
}

type CreateProfile struct {
	FirstName   string    `json:"first_name"`
	MiddleName  string    `json:"middle_name,omitempty"` // omitempty makes it optional in the schema
	LastName    string    `json:"last_name"`
	DateOfBirth time.Time `json:"date_of_birth"`
	Gender      string    `json:"gender"`
	Country     string    `json:"country"`
	Email       string    `json:"email"`
	Phone       string    `json:"phone,omitempty"`
	Password    string    `json:"password"`
	PIN         string    `json:"pin,omitempty"`
}

type LoginResponse struct {
	AuthorizationID string
	Profile         *json.RawMessage
}

type Profile struct {
	FirstName        string    `json:"first_name"`
	MiddleName       string    `json:"middle_name,omitempty"` // omitempty makes it optional in the schema
	LastName         string    `json:"last_name"`
	DateOfBirth      time.Time `json:"date_of_birth"`
	Gender           string    `json:"gender"`
	Country          string    `json:"country"`
	Email            string    `json:"email"`
	Phone            string    `json:"phone,omitempty"`
	PhoneVerified    bool
	EmailVerified    bool
	IdentityVerified string
}

func (p *Profile) Filter(s sqlc.User) error {
	p.FirstName = s.FirstName
	p.LastName = s.LastName
	p.Gender = s.Gender
	p.Country = s.Country
	p.Email = s.Email

	if s.MiddleName.Valid {
		p.MiddleName = s.MiddleName.String
	}

	if s.DateOfBirth.Valid {
		p.DateOfBirth = s.DateOfBirth.Time
	}

	if s.Phone.Valid {
		p.Phone = s.Phone.String
	}

	if s.IsPhoneVerified.Valid {
		p.PhoneVerified = s.IsPhoneVerified.Bool
	}

	if s.IsEmailVerified.Valid {
		p.EmailVerified = s.IsEmailVerified.Bool
	}
	p.IdentityVerified = string(s.IsIdentityVerified)

	return nil
}

type LoginRequest struct {
	Mode       string              `json:"mode" validate:"required,oneof=email phone"`
	Identifier string              `json:"identifier" validate:"required"`
	Password   []byte              `json:"password" validate:"required,min=8"`
	ResultOpts *LoginReqResultOpts `json:"resultOpts,omitempty" validate:"omitempty"`
}

type LoginReqResultOpts struct {
	// 0 - none, 1 - return verified credentials, 2 - return unverified credentials
	VerifiedCredentials uint8    `json:"verifiedCredentials,omitempty" validate:"omitempty,oneof=0 1 2"`
	Schema              []string `json:"schema,omitempty" validate:"omitempty,dive,required"`
}

// FilterProfile receives a UserProfile and returns a map of only the requested fields based on the Schema.
func (opts *LoginReqResultOpts) FilterProfile(profile Profile) (map[string]interface{}, error) {

	if len(opts.Schema) == 0 {
		return nil, nil
	}
	return filter.Select(opts.Schema, profile)

	// for _, field := range opts.Schema {
	// 	switch field {
	// 	case "first_name":
	// 		result["first_name"] = profile.FirstName
	// 	case "middle_name":
	// 		result["middle_name"] = profile.MiddleName
	// 	case "last_name":
	// 		result["last_name"] = profile.LastName
	// 	case "date_of_birth":
	// 		result["date_of_birth"] = profile.DateOfBirth
	// 	case "gender":
	// 		result["gender"] = profile.Gender
	// 	case "country":
	// 		result["country"] = profile.Country
	// 	case "email":
	// 		result["email"] = profile.Email
	// 	case "phone":
	// 		result["phone"] = profile.Phone
	// 	case "is_phone_verified":
	// 		result["is_phone_verified"] = profile.PhoneVerified
	// 	case "is_email_verified":
	// 		result["is_email_verified"] = profile.EmailVerified
	// 	case "is_identity_verified":
	// 		result["is_identity_verified"] = profile.IdentityVerified
	// 	}
	// }
}
