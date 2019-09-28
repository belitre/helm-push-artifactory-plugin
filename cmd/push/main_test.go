package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"

	"helm.sh/helm/pkg/cli"
	"helm.sh/helm/pkg/getter"
	"helm.sh/helm/pkg/repo"
)

var (
	settings           = &cli.EnvSettings{}
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

	// Create new Helm home w/ test repo
	tmp, err := ioutil.TempDir("", "helm-push-test")
	if err != nil {
		t.Error("unexpected error creating temp test dir", err)
	}
	defer os.RemoveAll(tmp)

	filepath := path.Join(tmp, "repositories.yaml")

	f := repo.NewFile()

	entry := repo.Entry{}
	entry.Name = "helm-push-test"
	entry.URL = ts.URL

	_, err = repo.NewChartRepository(&entry, getter.All(settings))
	if err != nil {
		t.Error("unexpected error created test repository", err)
	}

	f.Update(&entry)
	err = f.WriteFile(filepath, 0644)
	assert.NoError(t, err)

	os.Setenv("HELM_REPOSITORY_CONFIG", filepath)
	os.Setenv("HELM_REPO_USERNAME", "myuser")
	os.Setenv("HELM_REPO_PASSWORD", "mypass")

	// Not enough args
	args := []string{}
	cmd, err := newPushCmd(args)
	assert.NoError(t, err)
	err = cmd.RunE(cmd, args)
	assert.Error(t, err)

	// Bad chart path
	args = []string{"/this/not/a/chart", "helm-push-test"}
	cmd, err = newPushCmd(args)
	assert.NoError(t, err)
	err = cmd.RunE(cmd, args)
	assert.Error(t, err)

	// Bad repo name
	args = []string{testTarballPath, "wkerjbnkwejrnkj"}
	cmd, err = newPushCmd(args)
	assert.NoError(t, err)
	err = cmd.RunE(cmd, args)
	assert.Error(t, err)

	// Happy path
	args = []string{testTarballPath, "helm-push-test"}
	cmd, err = newPushCmd(args)
	assert.NoError(t, err)
	err = cmd.RunE(cmd, args)
	assert.NoError(t, err)

	// Happy path by repo URL
	args = []string{testTarballPath, ts.URL}
	cmd, err = newPushCmd(args)
	assert.NoError(t, err)
	err = cmd.RunE(cmd, args)
	assert.NoError(t, err)

	// Trigger reindex error
	postStatusCode = 403
	body = "{\"errors\": [{\"message\": \"Error\", \"status\": 403}]}"
	args = []string{testTarballPath, ts.URL}
	cmd, err = newPushCmd(args)
	assert.NoError(t, err)
	err = cmd.RunE(cmd, args)
	assert.Error(t, err, "403: Error")

	// Trigger 409
	putStatusCode = 409
	body = "{\"errors\": [{\"message\": \"Error\", \"status\": 409}]}"
	args = []string{testTarballPath, ts.URL}
	cmd, err = newPushCmd(args)
	assert.NoError(t, err)
	err = cmd.RunE(cmd, args)
	assert.Error(t, err, "409: Error")

	// Unable to parse JSON response body
	putStatusCode = 500
	body = "qkewjrnvqejrnbvjern"
	args = []string{testTarballPath, ts.URL}
	cmd, err = newPushCmd(args)
	assert.NoError(t, err)
	err = cmd.RunE(cmd, args)
	assert.Error(t, err, "500: could not properly parse response JSON: qkewjrnvqejrnbvjern")

}

// update the expired tests certificates some day...
/*
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

	// Create new Helm home w/ test repo
	tmp, err := ioutil.TempDir("", "helm-push-test")
	if err != nil {
		t.Error("unexpected error creating temp test dir", err)
	}
	defer os.RemoveAll(tmp)

	home := helmpath.Home(tmp)
	f := repo.NewRepoFile()

	entry := repo.Entry{}
	entry.Name = "helm-push-test"
	entry.URL = ts.URL

	_, err = repo.NewChartRepository(&entry, getter.All(settings))
	if err != nil {
		t.Error("unexpected error created test repository", err)
	}

	f.Update(&entry)
	os.MkdirAll(home.Repository(), 0777)
	f.WriteFile(home.RepositoryFile(), 0644)

	os.Setenv("HELM_REPOSITORY_CONFIG", home.String())
	os.Setenv("HELM_REPO_USERNAME", "myuser")
	os.Setenv("HELM_REPO_PASSWORD", "mypass")

	//no certificate options
	args := []string{testTarballPath, "helm-push-test"}
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
*/
