package httpclient

import "net/http"

//go:generate mockgen -source=httpclient.go -destination=mocks/mock_httpclient.go -package=mocks

// HTTPClient is the minimal surface of *http.Client used by services to call
// other Verancial microservices synchronously. Depending on this interface
// instead of *http.Client directly lets callers inject a fake in tests
// instead of hitting the network.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// DefaultHTTPClient wraps *http.Client so it satisfies HTTPClient.
type DefaultHTTPClient struct {
	*http.Client
}

func New(client *http.Client) HTTPClient {
	return &DefaultHTTPClient{Client: client}
}
