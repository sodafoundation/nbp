package opensds

import (
	"log"

	"github.com/opensds/opensds/client"
	"github.com/opensds/opensds/pkg/utils/constants"
)

// GetClient return OpenSDS Client
func GetClient(endpoint string, authStrategy string) (*client.Client, error) {

	log.Printf("current OpenSDS client endpoint: %s, auth strategy: %s ", endpoint, authStrategy)

	cfg := &client.Config{Endpoint: endpoint}

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
