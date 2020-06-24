package helm

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/strvals"
	"sigs.k8s.io/yaml"
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
	err := chartutil.SaveDir(c.Chart, outDir)
	if err != nil {
		return "", fmt.Errorf("Error while saving chart: %s", err)
	}
	const ValuesfileName = "values.yaml"
	vf := filepath.Join(outDir, c.Name(), ValuesfileName)
	valuesMap, err := yaml.Marshal(c.Values)
	if err != nil {
		return "", fmt.Errorf("Couldn't read values file as YAML: %s", err)
	}
	err = ioutil.WriteFile(vf, valuesMap, 0644)
	if err != nil {
		return "", fmt.Errorf("Couldn't wring values file: %s", err)
	}
	chart, err := loader.LoadDir(filepath.Join(outDir, c.Name()))
	if err != nil {
		return "", fmt.Errorf("New chart with the values seems to be invalid (unable to load): %s", err)
	}
	return chartutil.Save(chart, outDir)
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
