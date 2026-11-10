package users

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/pat3icki/pennychoice/internal/db"
	"github.com/pat3icki/pennychoice/pkg/jschema"
	testdb "github.com/pat3icki/pennychoice/tests"
)

var (
	GlobalTestManager *Service
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	envs, db_connect_type := testdb.LoadTestEnv()
	caCertPEM := envs["TEST_CA_CERT_PEM"]
	dbUrl := envs["TEST_DATABASE_URL"]

	cfg := db.Config{
		URI: dbUrl,
		SSLRootCAData: func() []byte {
			if db_connect_type == "remote" {
				return []byte(caCertPEM)
			}
			return nil
		}(), // Raw CA certificate data
		MaxConnections: 20,
		MinConnections: 5,
	}

	client, err := db.NewClient(ctx, cfg, nil)
	if err != nil {
		log.Fatal(err)
	}
	schema, err := jschema.Generate[CreateProfile]()
	if err != nil {
		log.Fatal(err)
	}
	minLen1 := 1
	minLen7 := 7
	for _, prop := range []string{"first_name", "last_name", "email", "country", "gender"} {
		if p, ok := schema.Properties[prop]; ok {
			p.MinLength = &minLen1
		}
	}
	schema.Properties["password"] = &jsonschema.Schema{
		Type:      "string",
		MinLength: &minLen7,
	}
	jschema.LoadSchema[*jsonschema.Schema]("create_user.schema", schema)

	contestantSchema, err := jschema.Generate[CreateContestantProfile]()
	if err != nil {
		log.Fatal(err)
	}
	for _, prop := range []string{"user_id", "profile_name", "display_name"} {
		if p, ok := contestantSchema.Properties[prop]; ok {
			p.MinLength = &minLen1
		}
	}
	jschema.LoadSchema[*jsonschema.Schema]("create_contestant.schema", contestantSchema)

	GlobalTestManager = &Service{
		Postgres:    client,
		SchemanName: "create_user.schema",
	}
	os.Exit(m.Run())
}
