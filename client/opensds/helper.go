package opensds

import (
	"log"
	"os"
	"sync"

	"github.com/opensds/opensds/client"
	"github.com/opensds/opensds/pkg/utils/constants"
)

const (

	// OpenSDSEndPoint environment variable name
	OpenSDSEndPoint = "OPENSDS_ENDPOINT"

	// OpenSDSAuthStrategy environment variable name
	OpenSDSAuthStrategy = "OPENSDS_AUTH_STRATEGY"
)

var (
	opensdsClient *client.Client
)

// Concurrent security
var once sync.Once

// GetClient return OpenSDS Client
func GetClient(endpoint string, authStrategy string) *client.Client {
	once.Do(func() {
		if endpoint == "" {
			// Get endpoint from environment
			endpoint = os.Getenv(OpenSDSEndPoint)
			log.Printf("current OpenSDS Client endpoint: %s", endpoint)
		}

		if endpoint == "" {
			// Using default endpoint
			endpoint = constants.DefaultOpensdsEndpoint
			log.Printf("using default OpenSDS Client endpoint: %s", endpoint)
		}

		if authStrategy == "" {
			// Get auth strategy from environment
			authStrategy = os.Getenv(OpenSDSAuthStrategy)
			log.Printf("current OpenSDS Client auth strategy: %s", authStrategy)
		}

		if authStrategy == "" {
			// Using default auth strategy
			authStrategy = constants.Noauth
			log.Printf("using default OpenSDS Client auth strategy: %s", authStrategy)
		}

		cfg := &client.Config{Endpoint: endpoint}

		switch authStrategy {
		case client.Keystone:
			cfg.AuthOptions = client.LoadKeystoneAuthOptionsFromEnv()
		case client.Noauth:
			cfg.AuthOptions = client.LoadNoAuthOptionsFromEnv()
		default:
			cfg.AuthOptions = client.NewNoauthOptions(constants.DefaultTenantId)
		}

		opensdsClient = client.NewClient(cfg)
	})

	return opensdsClient
}
