package helm

import (
	"fmt"

	yaml "gopkg.in/yaml.v2"
	"k8s.io/helm/pkg/chartutil"
	cpb "k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/strvals"
)

type (
	// Chart is a helm package that contains metadata
	Chart struct {
		*cpb.Chart
	}
)

// SetVersion overrides the chart version
func (c *Chart) SetVersion(version string) {
	c.Metadata.Version = version
}

// SetAppVersion overrides the chart version
func (c *Chart) SetAppVersion(appVersion string) {
	c.Metadata.AppVersion = appVersion
}

// GetChartByName returns a chart by "name", which can be
// either a directory or .tgz package
func GetChartByName(name string) (*Chart, error) {
	cc, err := chartutil.Load(name)
	if err != nil {
		return nil, err
	}
	return &Chart{cc}, nil
}

// CreateChartPackage creates a new .tgz package in directory
func CreateChartPackage(c *Chart, outDir string) (string, error) {
	return chartutil.Save(c.Chart, outDir)
}

// OverrideValues overrides values in chart values.yaml file
func (c *Chart) OverrideValues(overrides []string) error {
	ovMap := map[string]interface{}{}

	for _, o := range overrides {
		if err := strvals.ParseInto(o, ovMap); err != nil {
			return fmt.Errorf("failed parsing --set data: %s", err)
		}
	}

	ovAsBytes, err := yaml.Marshal(ovMap)
	if err != nil {
		return fmt.Errorf("error while marshal values: %s", err)
	}

	cvals, err := chartutil.CoalesceValues(c.Chart, &cpb.Config{Raw: string(ovAsBytes)})
	if err != nil {
		return fmt.Errorf("error while overriding chart values: %s", err)
	}

	cvalsAsYaml, err := cvals.YAML()
	if err != nil {
		return fmt.Errorf("error parsing values to yaml: %s", err)
	}

	c.Values = &cpb.Config{Raw: cvalsAsYaml}
	return nil
}
