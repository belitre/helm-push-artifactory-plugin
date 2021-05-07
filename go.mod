module github.com/belitre/helm-push-artifactory-plugin

go 1.15

require (
	github.com/jfrog/jfrog-client-go v0.8.1
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v0.0.6
	github.com/stretchr/testify v1.5.1
	helm.sh/helm/v3 v3.1.1
	k8s.io/apimachinery v0.17.3 // indirect
	sigs.k8s.io/yaml v1.1.0
)
