package external

import (
	coreGateway "github.com/TCP404/OneTiny-core/gateway"
	coreClient "github.com/TCP404/OneTiny-core/gateway/client"
)

var Core = getCoreProcessClient()

func getCoreProcessClient() coreClient.Client {
	return coreGateway.Client(coreGateway.Process)
}

// func getCoreHTTPClient() coreClient.Client { return coreGateway.Client(coreGateway.HTTP) }

// func getCoreRPCClient() coreClient.Client { return coreGateway.Client(coreGateway.RPC) }
