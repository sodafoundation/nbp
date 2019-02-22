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

	// Noauth
	Noauth = "noauth"
)

	// Noauth
	Noauth = "noauth"
)

// Concurrent security
var once sync.Once

// GetClient return OpenSDS Client
func GetClient(endpoint string, authStrategy string) (*client.Client, error) {
		if endpoint == "" {
			// Using default endpoint
			endpoint = constants.DefaultOpensdsEndpoint
			log.Printf("using default OpenSDS Client endpoint: %s", endpoint)
		}

	if endpoint == "" {
		// Using default endpoint
		endpoint = constants.DefaultOpensdsEndpoint
		log.Printf("using default OpenSDS Client endpoint: %s", endpoint)
	}

		if authStrategy == "" {
			// Using default auth strategy
			authStrategy = Noauth
			log.Printf("using default OpenSDS Client auth strategy: %s", authStrategy)
		}

	if authStrategy == "" {
		// Using default auth strategy
		authStrategy = Noauth
		log.Printf("using default OpenSDS Client auth strategy: %s", authStrategy)
	}

		switch authStrategy {
		case client.Keystone:
			cfg.AuthOptions = client.LoadKeystoneAuthOptionsFromEnv()
		case client.Noauth:
			cfg.AuthOptions = client.LoadNoAuthOptionsFromEnv()
		default:
			cfg.AuthOptions = client.NewNoauthOptions(constants.DefaultTenantId)
		}

	var authOptions client.AuthOptions
	var err error

	switch authStrategy {
	case client.Keystone:
		authOptions, err = client.LoadKeystoneAuthOptionsFromEnv()
		if err != nil {
			return nil, err
		}
	case client.Noauth:
		authOptions = client.LoadNoAuthOptionsFromEnv()
	default:
		authOptions = client.NewNoauthOptions(constants.DefaultTenantId)
	}

	cfg.AuthOptions = authOptions

	return client.NewClient(cfg)
}
