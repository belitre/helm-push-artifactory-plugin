# Helm push artifactory plugin

__WARNING: THIS VERSION IS FOR HELM 3, FOR HELM 2 PLEASE USE VERSION [v0.4.0](https://github.com/belitre/helm-push-artifactory-plugin/releases/tag/v0.4.0)__

A Helm plugin to push helm charts to artifactory:
 
 * A version for artifactory of helm-push: https://github.com/chartmuseum/helm-push
 * Using a couple of things from Jfrog-cli-go: https://github.com/jfrog/jfrog-cli-go
 * And a bit of makefile magic from: https://github.com/helm/helm

## Install

To install the version for Helm 3 you need to specify the version when installing using helm cli:

```
$ helm plugin install https://github.com/belitre/helm-push-artifactory-plugin --version v1.0.0
Downloading and installing helm-push-artifactory v1.0.0 ...
https://github.com/belitre/helm-push-artifactory-plugin/releases/download/v1.0.0/helm-push-artifactory_v1.0.0_darwin_amd64.tar.gz
Installed plugin: push-artifactory
```

You can also download on one of the compressed files from [here](https://github.com/belitre/helm-push-artifactory-plugin/releases/tag/v1.0.0) and just extract it in your `$HELM_HOME/plugins/`

__Important for windows users: I really don't know how to make the `helm plugin install` command work on Windows :D so please just download the zip and extract it on your `$HELM_HOME/plugins/ folder :)__

## Uninstall

```
helm plugin remove push-artifactory
Removed plugin: push-artifactory
```

## Local and virtual repositories

Artifactory has two types of repositories: local and virtual. Local repositories are the ones where you push the charts, but to get a chart you'll need to use a virtual repository!

__This plugin works with local repositories__, you can add them through the Helm CLI like a virtual repository and use it later instead of the URL. But remember: __you won't be able to get charts from a local repository__

Example:

* We can add our local repository with helm CLI:

    ```bash
    $ helm repo add --username myuser --password mypass my-local-repo https://artifactoryhost/my-local-repo
    "my-local-repo" has been added to your repositories
    ```

* We can use this repository later to push charts:

    ```bash
    $ helm push-artifactory mychart-0.3.2.tgz my-local-repo 
    Pushing mychart-0.3.2.tgz to https://artifactoryhost/my-local-repo/mychart/mychart-0.3.2.tgz...
    Done.
    Reindex helm repository my-local-repo...
    Reindex of helm repo my-local-repo was scheduled to run.
    ```

* __We can't get the helm chart from a local repo:__

    ```bash
    $ helm fetch my-local-repo/mychart
    Error: Get local://mychart/mychart-0.3.2.tgz: unsupported protocol scheme "local"
    ```

* We can add the virtual repo and get the chart:

    ```bash
    $ helm repo add --username myuser --password mypass my-virtual-repo https://artifactoryhost/my-virtual-repo
    "my-virtual-repo" has been added to your repositories
    $ helm repo update
    Hang tight while we grab the latest from your chart repositories...
    ...Skip local chart repository
    ...Successfully got an update from the "my-local-repo" chart repository
    ...Successfully got an update from the "my-virtual-repo" chart repository
    Update Complete. ⎈ Happy Helming!⎈ 
    $ helm fetch my-virtual-repo/mychart
    $ ls
    mychart-0.3.2.tgz
    ``` 

## Usage

You can override the repositories file using the environment variable `HELM_REPOSITORY_CONFIG`, example:

```
export HELM_REPOSITORY_CONFIG=/home/myuser/my_repositories.yaml
```

Example using URL:

```bash
$ helm push-artifactory /my/chart/folder https://my-artifactory/my-local-repo --username username --password password
```

Example using helm repo added through CLI:
```bash
$ helm push-artifactory /my/chart/folder my-local-repo
```

For all available plugin options, please run:
```bash
$ helm push-artifactory --help
```

### Pushing a directory
Point to a directory containing a valid `Chart.yaml` and the chart will be packaged and uploaded:
```bash
$ cat mychart/Chart.yaml
name: mychart
version: 0.3.2
```
```bash
$ helm push-artifactory mychart/ https://my-artifactory/my-local-repo
Pushing mychart-0.3.2.tgz to https://my-artifactory/my-local-repo/mychart/mychart-0.3.2.tgz...
Done.
Reindex helm repository my-local-repo...
Reindex of helm repo my-local-repo was scheduled to run.
```

### Pushing with a custom version
The `--version` or `-v` flag can be provided, which will push the package with a custom version.

Here is an example using the last git commit id as the version:
```bash
$ helm push-artifactory mychart/ --version="$(git log -1 --pretty=format:%h)" https://my-artifactory/my-local-repo
Pushing mychart-5abbbf28.tgz to https://my-artifactory/my-local-repo/mychart/mychart-5abbbf28.tgz...
Done.
Reindex helm repository my-local-repo...
Reindex of helm repo my-local-repo was scheduled to run.
```

### Pushing overriding values
The `--set` or `-s` flag can be provided, which will push the package overriding values in the values.yaml file.

Here is an example using a custom version and default image tag:
```bash
$ helm push-artifactory mychart/ -v="0.5.1" -s image.tag="0.5.1" https://my-artifactory/my-local-repo
Pushing mychart-0.5.1.tgz to https://my-artifactory/my-local-repo/mychart/mychart-0.5.1.tgz...
Done.
Reindex helm repository my-local-repo...
Reindex of helm repo my-local-repo was scheduled to run.
```

If we check the values.yaml of the chart we'll find:
```
image:
  tag: 0.5.1
```

We can override multiple values:
```bash
$ helm push-artifactory mychart/ -v="0.5.1" -s image.tag="0.5.1" -s service.name="my-custom-name" https://my-artifactory/my-local-repo
Pushing mychart-0.5.1.tgz to https://my-artifactory/my-local-repo/mychart/mychart-0.5.1.tgz...
Done.
Reindex helm repository my-local-repo...
Reindex of helm repo my-local-repo was scheduled to run.
```

This will set in values.yaml:
```
image:
  tag: 0.5.1
service:
  name: my-custom-name
```

### Push .tgz package
This workflow does not require the use of `helm package`, but pushing .tgz is still supported:
```bash
$ helm push-artifactory mychart-0.3.2.tgz https://my-artifactory/my-local-repo
Pushing mychart-0.3.2.tgz to https://my-artifactory/my-local-repo/mychart/mychart-0.3.2.tgz...
Done.
Reindex helm repository my-local-repo...
Reindex of helm repo my-local-repo was scheduled to run.
```

### Push with path
You can set a path to push your chart in your Artifactory local repository:
```bash
$ helm push-artifactory mychart/ https://my-artifactory/my-local-repo --path organization
Pushing mychart-0.3.2.tgz to https://my-artifactory/my-local-repo/organization/mychart/mychart-0.3.2.tgz...
Done.
Reindex helm repository my-local-repo...
Reindex of helm repo my-local-repo was scheduled to run.
```

### Skip repository reindex
You can skip triggering the repository reindex:
```bash
$ helm push-artifactory mychart/ https://my-artifactory/my-local-repo --skip-reindex
Pushing mychart-0.3.2.tgz to https://my-artifactory/my-local-repo/mychart/mychart-0.3.2.tgz...
Done.
```

## Authentication
### Basic Auth
__The plugin will not use the auth info located in `~/.helm/repository/repositories.yaml` in order to authenticate.__

You can provide username and password through commmand line with `--username username --password password` or use the following environment variables for basic auth on push operations:
```bash
$ export HELM_REPO_USERNAME="myuser"
$ export HELM_REPO_PASSWORD="mypass"
```

### Access Token
You can provide an access token through command line with `--access-token my-token` or use the following env var:
```bash
$ export HELM_REPO_ACCESS_TOKEN="<token>"
```

If only the access token is supplied without any username, the plugin will send the token in the header:
```
Authorization: Bearer <token>
```

If a username is supplied with an access token, the plugin will use basic authentication, using the access token as password for the user.

### Api Key
You can provide an api key through command line with `--api-key my-key` or use the following env var:
```bash
$ export HELM_REPO_API_KEY="<api-key>"
```

If only the api key is supplied without any username, the plugin will send the api key in the header:
```
X-JFrog-Art-Api: <api-key>
```

If a username is supplied with an api key, the plugin will use basic authentication, using the api key as password for the user.

### TLS Client Cert Auth

If you need to setup your TLS cert authentication, the following options are available:

```
--ca-file string    Verify certificates of HTTPS-enabled servers using this CA bundle [$HELM_REPO_CA_FILE]
--cert-file string  Identify HTTPS client using this SSL certificate file [$HELM_REPO_CERT_FILE]
--key-file string   Identify HTTPS client using this SSL key file [$HELM_REPO_KEY_FILE]
--insecure          Connect to server with an insecure way by skipping certificate verification [$HELM_REPO_INSECURE]
```

