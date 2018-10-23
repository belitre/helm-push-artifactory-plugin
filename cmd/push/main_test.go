package main

import (
	"crypto/rand"
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	helm_env "k8s.io/helm/pkg/helm/environment"
	"k8s.io/helm/pkg/tlsutil"
)

var (
	settings           helm_env.EnvSettings
	testTarballPath    = "../../testdata/charts/mychart/mychart-0.1.0.tgz"
	testCertPath       = "../../testdata/tls/test_cert.crt"
	testKeyPath        = "../../testdata/tls/test_key.key"
	testCAPath         = "../../testdata/tls/ca.crt"
	testServerCAPath   = "../../testdata/tls/server_ca.crt"
	testServerCertPath = "../../testdata/tls/test_server.crt"
	testServerKeyPath  = "../../testdata/tls/test_server.key"
)

func TestPushCmd(t *testing.T) {
	postStatusCode := 200
	putStatusCode := 201
	body := "Just a message"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			w.WriteHeader(postStatusCode)
		case "PUT":
			w.WriteHeader(putStatusCode)
		default:
			w.WriteHeader(404)
		}
		w.Write([]byte(body))
	}))
	defer ts.Close()

	os.Setenv("HELM_REPO_USERNAME", "myuser")
	os.Setenv("HELM_REPO_PASSWORD", "mypass")

	// Not enough args
	args := []string{}
	cmd := newPushCmd(args)
	err := cmd.RunE(cmd, args)
	if err == nil {
		t.Error("expecting error with missing args, instead got nil")
	}

	// Bad chart path
	args = []string{"/this/this/not/a/chart", ts.URL}
	cmd = newPushCmd(args)
	err = cmd.RunE(cmd, args)
	if err == nil {
		t.Error("expecting error with bad chart path, instead got nil")
	}

	// Bad repo URL
	args = []string{testTarballPath, "wkerjbnkwejrnkj"}
	cmd = newPushCmd(args)
	err = cmd.RunE(cmd, args)
	if err == nil {
		t.Error("expecting error with bad repo URL, instead got nil")
	}

	// Happy path
	args = []string{testTarballPath, ts.URL}
	cmd = newPushCmd(args)
	err = cmd.RunE(cmd, args)
	if err != nil {
		t.Error("unexpecting error uploading tarball", err)
	}

	// Trigger reindex error
	postStatusCode = 403
	body = "{\"errors\": [{\"message\": \"Error\", \"status\": 403}]}"
	args = []string{testTarballPath, ts.URL}
	cmd = newPushCmd(args)
	err = cmd.RunE(cmd, args)
	if err == nil {
		t.Error("expecting error with 403, instead got nil")
	} else {
		assert.Error(t, err, "403: Error")
	}

	// Trigger 409
	putStatusCode = 409
	body = "{\"errors\": [{\"message\": \"Error\", \"status\": 409}]}"
	args = []string{testTarballPath, ts.URL}
	cmd = newPushCmd(args)
	err = cmd.RunE(cmd, args)
	if err == nil {
		t.Error("expecting error with 409, instead got nil")
	} else {
		assert.Error(t, err, "409: Error")
	}

	// Unable to parse JSON response body
	putStatusCode = 500
	body = "qkewjrnvqejrnbvjern"
	args = []string{testTarballPath, ts.URL}
	cmd = newPushCmd(args)
	err = cmd.RunE(cmd, args)
	if err == nil {
		t.Error("expecting error with bad response body, instead got nil")
	} else {
		assert.Error(t, err, "500: could not properly parse response JSON: qkewjrnvqejrnbvjern")
	}

}

func TestPushCmdWithTlsEnabledServer(t *testing.T) {
	body := "Just a message."
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			w.WriteHeader(200)
		}
		if r.Method == "PUT" {
			w.WriteHeader(201)
		}
		w.Write([]byte(body))
	}))
	cert, err := tls.LoadX509KeyPair(testCertPath, testKeyPath)
	if err != nil {
		t.Fatalf("failed to load certificate and key with error: %s", err.Error())
	}

	caCertPool, err := tlsutil.CertPoolFromFile(testServerCAPath)
	if err != nil {
		t.Fatalf("load server CA file failed with error: %s", err.Error())
	}

	ts.TLS = &tls.Config{
		ClientCAs:    caCertPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{cert},
		Rand:         rand.Reader,
	}
	ts.StartTLS()
	defer ts.Close()

	os.Setenv("HELM_REPO_USERNAME", "myuser")
	os.Setenv("HELM_REPO_PASSWORD", "mypass")

	//no certificate options
	args := []string{testTarballPath, ts.URL}
	cmd := newPushCmd(args)
	err = cmd.RunE(cmd, args)
	if err == nil {
		t.Fatal("expected non nil error but got nil when run cmd without certificate option")
	}

	os.Setenv("HELM_REPO_CA_FILE", testCAPath)
	os.Setenv("HELM_REPO_CERT_FILE", testServerCertPath)
	os.Setenv("HELM_REPO_KEY_FILE", testServerKeyPath)

	err = cmd.RunE(cmd, args)
	if err != nil {
		t.Fatalf("unexpecting error uploading tarball: %s", err)
	}
}
