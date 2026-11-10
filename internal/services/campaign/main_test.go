package campaign

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
	schema, err := jschema.Generate[CreateCampaignV1]()
	if err != nil {
		log.Fatal(err)
	}
	jschema.LoadSchema[*jsonschema.Schema]("create_campaign_v1.schema", schema)

	GlobalTestManager = &Service{
		Postgres:    client,
		SchemanName: "create_campaign_v1.schema",
	}
	os.Exit(m.Run())
}
