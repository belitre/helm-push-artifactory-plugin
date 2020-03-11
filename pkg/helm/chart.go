package helm

import (
	"fmt"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/strvals"
)

type (
	// Chart is a helm package that contains metadata
	Chart struct {
		*chart.Chart
	}
)

// SetVersion overrides the chart version
func (c *Chart) SetVersion(version string) {
	c.Metadata.Version = version
}

// GetChartByName returns a chart by "name", which can be
// either a directory or .tgz package
func GetChartByName(name string) (*Chart, error) {
	chartLoader, err := loader.Loader(name)
	if err != nil {
		return nil, err
	}
	cc, err := chartLoader.Load()
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

	cvals, err := chartutil.CoalesceValues(c.Chart, ovMap)
	if err != nil {
		return fmt.Errorf("Error while overriding chart values: %s", err)
	}

	c.Values = cvals
	return nil
}
