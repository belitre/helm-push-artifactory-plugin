package artifactory

import (
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	cmClient, err := NewClient(
		URL("http://localhost:8080"),
		Path("/my/path"),
		Username("user"),
		Password("pass"),
		AccessToken("accessToken"),
		ApiKey("apiKey"),
		Timeout(60),
		CAFile("../../testdata/tls/ca.crt"),
		KeyFile("../../testdata/tls/test_key.key"),
		CertFile("../../testdata/tls/test_cert.crt"),
		InsecureSkipVerify(true),
	)

	if err != nil {
		t.Fatalf("expect creating a client instance but met error: %s", err)
	}

	c, ok := cmClient.(*client)

	if !ok {
		t.Fatalf("incorrect type for client")
	}

	if c.opts.url != "http://localhost:8080" {
		t.Errorf("expected url to be http://localhost:8080, got %v", c.opts.url)
	}

	if c.opts.username != "user" {
		t.Errorf("expected username to be user, got %v", c.opts.username)
	}

	if c.opts.password != "pass" {
		t.Errorf("expected password to be pass, got %v", c.opts.password)
	}

	if c.opts.accessToken != "accessToken" {
		t.Errorf("expected accessToken to be accessToken, got %v", c.opts.accessToken)
	}

	if c.opts.apiKey != "apiKey" {
		t.Errorf("expected apiKey to be apiKey, got %v", c.opts.apiKey)
	}

	if c.opts.timeout != time.Minute {
		t.Errorf("expected timeout duration to be 1 minute, got %v", c.opts.timeout)
	}

	if c.opts.caFile != "../../testdata/tls/ca.crt" {
		t.Errorf("expected ca file path to be '../../testdata/tls/ca.crt' but got %v", c.opts.caFile)
	}

	if c.opts.certFile != "../../testdata/tls/test_cert.crt" {
		t.Errorf("expected cert file path to be '../../testdata/tls/test_cert.crt' but got %v", c.opts.certFile)
	}

	if c.opts.keyFile != "../../testdata/tls/test_key.key" {
		t.Errorf("expected key file path to be '../../testdata/tls/test_key.key' but got %v", c.opts.keyFile)
	}

	if !c.opts.insecureSkipVerify {
		t.Errorf("expected insecure flag to be 'true' but got %v", c.opts.insecureSkipVerify)
	}
}
