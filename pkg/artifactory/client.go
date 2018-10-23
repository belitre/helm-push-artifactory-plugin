package artifactory

import (
	"fmt"
	"net/http"

	"k8s.io/helm/pkg/tlsutil"
)

type (
	// Client is an HTTP client to connect to ChartMuseum
	client struct {
		*http.Client
		opts options
	}

	Client interface {
		UploadChartPackage(chartName, chartPackagePath string) (*http.Response, error)
		ReindexArtifactoryRepo() (*http.Response, error)
	}
)

// Option configures the client with the provided options.
func (c *client) Option(opts ...Option) *client {
	for _, opt := range opts {
		opt(&c.opts)
	}
	return c
}

// NewClient creates a new client.
func NewClient(opts ...Option) (Client, error) {
	var c client
	c.Client = &http.Client{}
	c.Option(Timeout(30))
	c.Option(opts...)
	c.Timeout = c.opts.timeout

	//Enable tls config if configured
	tr, err := newTransport(
		c.opts.certFile,
		c.opts.keyFile,
		c.opts.caFile,
		c.opts.insecureSkipVerify,
	)
	if err != nil {
		return nil, err
	}

	c.Transport = tr

	return &c, nil
}

//Create transport with TLS config
func newTransport(certFile, keyFile, caFile string, insecureSkipVerify bool) (*http.Transport, error) {
	transport := &http.Transport{}

	tlsConf, err := tlsutil.NewClientTLS(certFile, keyFile, caFile)
	if err != nil {
		return nil, fmt.Errorf("can't create TLS config: %s", err.Error())
	}

	tlsConf.InsecureSkipVerify = insecureSkipVerify

	transport.TLSClientConfig = tlsConf
	transport.Proxy = http.ProxyFromEnvironment

	return transport, nil
}
