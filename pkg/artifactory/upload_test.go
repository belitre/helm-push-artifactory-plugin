package artifactory

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	testTarballPath    = "../../testdata/charts/mychart/mychart-0.1.0.tgz"
	testCertPath       = "../../testdata/tls/test_cert.crt"
	testKeyPath        = "../../testdata/tls/test_key.key"
	testCAPath         = "../../testdata/tls/ca.crt"
	testServerCAPath   = "../../testdata/tls/server_ca.crt"
	testServerCertPath = "../../testdata/tls/test_server.crt"
	testServerKeyPath  = "../../testdata/tls/test_server.key"
)

func TestUploadChartPackage(t *testing.T) {
	chartName := "mychart"
	repoName := "myrepo"
	basicAuthHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte("user:pass"))
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("URL:  " + r.URL.String())
		if r.URL.String() != "/myrepo/my/path/mychart/mychart-0.1.0.tgz" {
			w.WriteHeader(404)
		} else if r.Header.Get("Authorization") != basicAuthHeader {
			w.WriteHeader(401)
		} else {
			w.WriteHeader(201)
		}
	}))
	defer ts.Close()

	url := ts.URL + "/" + repoName

	// Happy path
	cmClient, err := NewClient(
		URL(url),
		Username("user"),
		Password("pass"),
		Path("/my/path"),
	)
	if err != nil {
		t.Fatalf("[happy path] expect creating a client instance but met error: %s", err)
	}
	resp, err := cmClient.UploadChartPackage(chartName, testTarballPath)
	if err != nil {
		t.Error("error uploading chart package", err)
	}
	if resp.StatusCode != 201 {
		t.Errorf("expecting 201 instead got %d", resp.StatusCode)
	}

	// Bad package path
	resp, err = cmClient.UploadChartPackage(chartName, "/non/existant/path/mychart-0.1.0.tgz")
	if err == nil {
		t.Error("expecting error with bad package path, instead got nil")
	}

	// Bad URL
	cmClient, _ = NewClient(URL("jaswehfgew"))
	_, err = cmClient.UploadChartPackage(chartName, testTarballPath)
	if err == nil {
		t.Error("[bad URL] expecting error with bad package path, instead got nil")
	}

	// Bad context path
	cmClient, err = NewClient(
		URL(url),
		Username("user"),
		Password("pass"),
		Path("/my/crappy/context/path"),
		Timeout(5),
	)
	if err != nil {
		t.Fatalf("[bad path] expect creating a client instance but met error: %s", err)
	}

	resp, err = cmClient.UploadChartPackage(chartName, testTarballPath)
	if err != nil {
		t.Error("unexpected error with bad context path", err)
	}
	if resp.StatusCode != 404 {
		t.Errorf("expecting 404 instead got %d", resp.StatusCode)
	}

	// Unauthorized, invalid user/pass combo (basic auth)
	cmClient, err = NewClient(
		URL(url),
		Username("baduser"),
		Password("badpass"),
		Path("/my/path"),
	)
	if err != nil {
		t.Fatalf("[unauthorized: invalid user/pass] expect creating a client instance but met error: %s", err)
	}
	resp, err = cmClient.UploadChartPackage(chartName, testTarballPath)
	if err != nil {
		t.Error("unexpected error with invalid user/pass combo (basic auth)", err)
	}
	if resp.StatusCode != 401 {
		t.Errorf("expecting 401 instead got %d", resp.StatusCode)
	}

	// Unauthorized, missing user/pass combo (basic auth)
	cmClient, err = NewClient(
		URL(url),
		Path("/my/path"),
	)
	if err != nil {
		t.Fatalf("[unauthorized: missing user/pass] expect creating a client instance but met error: %s", err)
	}
	resp, err = cmClient.UploadChartPackage(chartName, testTarballPath)
	if err != nil {
		t.Error("unexpected error with missing user/pass combo (basic auth)", err)
	}
	if resp.StatusCode != 401 {
		t.Errorf("expecting 401 instead got %d", resp.StatusCode)
	}
}

/*
func TestUploadChartPackageWithTlsServer(t *testing.T) {
	chartName := "mychart"
	basicAuthHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte("user:pass"))
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.String(), "/my/path") {
			w.WriteHeader(404)
		} else if r.Header.Get("Authorization") != basicAuthHeader {
			w.WriteHeader(401)
		} else {
			w.WriteHeader(201)
		}
	}))

	cert, err := tls.LoadX509KeyPair(testCertPath, testKeyPath)
	if err != nil {
		t.Fatalf("failed to load certificate and key with error: %s", err.Error())
	}

	ts.TLS = &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	ts.StartTLS()
	defer ts.Close()

	//Without certificate
	cmClient, err := NewClient(
		URL(ts.URL),
		Username("user"),
		Password("pass"),
		Path("/my/path"),
	)
	if err != nil {
		t.Fatalf("[without certificate] expect creating a client instance but met error: %s", err)
	}

	_, err = cmClient.UploadChartPackage(chartName, testTarballPath)
	if err == nil {
		t.Fatal("expected error returned when uploading package without cert to tls enabled https server")
	}

	//Enable insecure flag
	cmClient, err = NewClient(
		URL(ts.URL),
		Username("user"),
		Password("pass"),
		Path("/my/path"),
		InsecureSkipVerify(true),
	)
	if err != nil {
		t.Fatalf("[enable insecure flag] expect creating a client instance but met error: %s", err)
	}

	resp, err := cmClient.UploadChartPackage(chartName, testTarballPath)
	if err != nil {
		t.Fatalf("[enable insecure flag] expected nil error but got %s", err.Error())
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("[enable insecure flag] expect status code 201 but got %d", resp.StatusCode)
	}

	//Upload with ca file
	cmClient, err = NewClient(
		URL(ts.URL),
		Username("user"),
		Password("pass"),
		Path("/my/path"),
		CAFile(testCAPath),
	)
	if err != nil {
		t.Fatalf("[upload with ca file] expect creating a client instance but met error: %s", err)
	}

	resp, err = cmClient.UploadChartPackage(chartName, testTarballPath)
	if err != nil {
		t.Fatalf("[upload with ca file] expected nil error but got %s", err.Error())
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("[upload with ca file] expect status code 201 but got %d", resp.StatusCode)
	}
}

func TestUploadChartPackageWithVerifyingClientCert(t *testing.T) {
	chartName := "mychart"
	basicAuthHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte("user:pass"))
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.String(), "/my/path") {
			w.WriteHeader(404)
		} else if r.Header.Get("Authorization") != basicAuthHeader {
			w.WriteHeader(401)
		} else {
			w.WriteHeader(201)
		}
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

	//Upload with cert and key files
	cmClient, err := NewClient(
		URL(ts.URL),
		Username("user"),
		Password("pass"),
		Path("/my/path"),
		KeyFile(testServerKeyPath),
		CertFile(testServerCertPath),
		CAFile(testCAPath),
	)
	if err != nil {
		t.Fatalf("[upload with cert and key files] expect creating a client instance but met error: %s", err)
	}

	resp, err := cmClient.UploadChartPackage(chartName, testTarballPath)
	if err != nil {
		t.Fatalf("[upload with cert and key files] expected nil error but got %s", err.Error())
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("[upload with cert and key files] expect status code 201 but got %d", resp.StatusCode)
	}
}
*/

func TestReindexArtifactoryRepo(t *testing.T) {
	basicAuthApiKey := "Basic " + base64.StdEncoding.EncodeToString([]byte("user:apiKey"))
	basicAuthToken := "Basic " + base64.StdEncoding.EncodeToString([]byte("user:token"))
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() != "/artifactory/api/helm/myrepo/reindex" {
			w.WriteHeader(404)
		} else if r.Header.Get("Authorization") != basicAuthApiKey &&
			r.Header.Get("Authorization") != basicAuthToken &&
			r.Header.Get("Authorization") != "Bearer token" &&
			r.Header.Get("X-JFrog-Art-Api") != "apiKey" {
			w.WriteHeader(401)
		} else {
			w.WriteHeader(200)
		}
	}))
	defer ts.Close()

	url := ts.URL + "/artifactory/myrepo"

	// basic auth token
	cmClient, err := NewClient(
		URL(url),
		Username("user"),
		AccessToken("token"),
		Path("/my/path"),
	)
	if err != nil {
		t.Fatalf("[basic auth token] expect creating a client instance but met error: %s", err)
	}
	resp, err := cmClient.ReindexArtifactoryRepo()
	if err != nil {
		t.Error("error reindexing repo", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expecting 200 instead got %d", resp.StatusCode)
	}

	// auth header token
	cmClient, err = NewClient(
		URL(url),
		AccessToken("token"),
	)
	if err != nil {
		t.Fatalf("[auth header token] expect creating a client instance but met error: %s", err)
	}
	resp, err = cmClient.ReindexArtifactoryRepo()
	if err != nil {
		t.Error("error reindexing repo", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expecting 200 instead got %d", resp.StatusCode)
	}

	// basic auth apiKey
	cmClient, err = NewClient(
		URL(url),
		Username("user"),
		ApiKey("apiKey"),
	)
	if err != nil {
		t.Fatalf("[basic auth apiKey] expect creating a client instance but met error: %s", err)
	}
	resp, err = cmClient.ReindexArtifactoryRepo()
	if err != nil {
		t.Error("error reindexing repo", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expecting 200 instead got %d", resp.StatusCode)
	}

	// auth header apiKey
	cmClient, err = NewClient(
		URL(url),
		ApiKey("apiKey"),
	)
	if err != nil {
		t.Fatalf("[auth header apiKey] expect creating a client instance but met error: %s", err)
	}
	resp, err = cmClient.ReindexArtifactoryRepo()
	if err != nil {
		t.Error("error reindexing repo", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expecting 200 instead got %d", resp.StatusCode)
	}
}
