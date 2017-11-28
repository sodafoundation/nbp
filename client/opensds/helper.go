package opensds

import (
	"log"
	"os"

	"github.com/opensds/opensds/client"
)

const (
	// OpenSDSEndPoint environment variable name
	OpenSDSEndPoint = "OPENSDS_ENDPOINT"
)

// GetClient return OpenSDS Client
func GetClient(endpoint string) *client.Client {

	if endpoint == "" {
		// Get endpoint from environment
		endpoint = os.Getenv(OpenSDSEndPoint)
		log.Printf("current OpenSDS Client endpoint: %s", endpoint)
	}

	if endpoint == "" {
		// Using default endpoint
		endpoint = "http://localhost:50040"
		log.Printf("using default OpenSDS Client endpoint: %s", endpoint)
	}

	return client.NewClient(
		&client.Config{
			Endpoint: endpoint,
		})
}

