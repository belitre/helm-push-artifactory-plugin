package artifactory

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/io/fileutils"
)

// ReindexArtifactoryRepo calculates chart index of the local repository
func (c *client) ReindexArtifactoryRepo() (*http.Response, error) {
	u, err := url.Parse(c.opts.url)
	if err != nil {
		return nil, err
	}

	artifactoryBase, repoName := path.Split(strings.TrimRight(u.Path, "/"))

	u.Path = path.Join(artifactoryBase, "api/helm/", repoName, "reindex")

	req, err := http.NewRequest("POST", u.String(), nil)
	if err != nil {
		return nil, err
	}

	c.addHeaders(req)

	fmt.Printf("Reindex helm repository %s...\n", repoName)
	return c.Do(req)
}

// UploadChartPackage uploads a chart package to Artifactory (PUT https://repoURL/path/chart.version.tgz)
func (c *client) UploadChartPackage(chartName, chartPackagePath string) (*http.Response, error) {
	f, err := os.Open(chartPackagePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	u, err := url.Parse(c.opts.url)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, c.opts.path, chartName, path.Base(chartPackagePath))
	req, err := buildRequest(u.String(), f)
	if err != nil {
		return nil, err
	}

	c.addHeaders(req)

	fmt.Printf("Pushing %s to %s...\n", filepath.Base(chartPackagePath), u.String())
	return c.Do(req)
}

func (c *client) addHeaders(req *http.Request) {
	if c.opts.apiKey != "" {
		if c.opts.username != "" {
			req.SetBasicAuth(c.opts.username, c.opts.apiKey)
		} else {
			req.Header.Set("X-JFrog-Art-Api", c.opts.apiKey)
		}
	} else if c.opts.password != "" {
		req.SetBasicAuth(c.opts.username, c.opts.password)
	} else if c.opts.accessToken != "" {
		if c.opts.username != "" {
			req.SetBasicAuth(c.opts.username, c.opts.accessToken)
		} else {
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.opts.accessToken))
		}
	}

	req.Header.Set("User-Agent", "helm-push-artifactory-plugin")
}

func buildRequest(url string, f *os.File) (*http.Request, error) {
	req, err := http.NewRequest("PUT", url, fileutils.GetUploadRequestContent(f))
	if err != nil {
		return nil, err
	}

	details, err := fileutils.GetFileDetails(f.Name())
	if err != nil {
		return nil, err
	}

	length := strconv.FormatInt(details.Size, 10)
	req.Header.Set("Content-Length", length)
	req.Header.Set("X-Checksum-Sha1", details.Checksum.Sha1)
	req.Header.Set("X-Checksum-Md5", details.Checksum.Md5)
	if len(details.Checksum.Sha256) > 0 {
		req.Header.Set("X-Checksum", details.Checksum.Sha256)
	}

	return req, nil
}
