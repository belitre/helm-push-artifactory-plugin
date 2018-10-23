package version

import (
	"fmt"
)

// Version should be updated each time there is a new release
var (
	Version   = "Canary"
	GitCommit = ""
)

func GetVersion() string {
	v := fmt.Sprintf("Version: %s", Version)
	if len(GitCommit) > 0 {
		v = v + fmt.Sprintf("-%s", GitCommit)
	}
	return v
}
