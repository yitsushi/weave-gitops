package fluxops

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/weaveworks/weave-gitops/pkg/status"
	"github.com/weaveworks/weave-gitops/pkg/utils"
	"github.com/weaveworks/weave-gitops/pkg/version"
	"sigs.k8s.io/yaml"
)

var (
	fluxHandler = defaultFluxHandler
	fluxBinary  string
)

func FluxPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	path := fmt.Sprintf("%v/.wego/bin", homeDir)
	return fmt.Sprintf("%v/flux-%v", path, version.FluxVersion), nil
}

func FluxBinaryPath() string {
	return fluxBinary
}

func SetFluxHandler(f func(string) ([]byte, error)) {
	fluxHandler = f
}

func CallFlux(arglist string) ([]byte, error) {
	return fluxHandler(arglist)
}

func defaultFluxHandler(arglist string) ([]byte, error) {
	initFluxBinary()
	return utils.CallCommand(fmt.Sprintf("%s %s", fluxBinary, arglist))
}

func Bootstrap(owner, repoName string) error {
	isOrg, err := isOrganization(owner)
	if err != nil {
		return err
	}
	if isOrg {
		if _, err := CallFlux(fmt.Sprintf("bootstrap github --owner=%s --repository=%s", owner, repoName)); err != nil {
			return err
		}
	} else {
		if _, err := CallFlux(fmt.Sprintf("bootstrap github --owner=%s --repository=%s --branch=main --private=false --personal=true", owner, repoName)); err != nil {
			return err
		}
	}
	if status.GetClusterStatus() != status.FluxInstalled {
		return fmt.Errorf("Failed to install flux")
	}

	return nil
}

func GetOwnerFromEnv() (string, error) {
	// check for github username
	user, okUser := os.LookupEnv("GITHUB_ORG")
	if okUser {
		return user, nil
	}

	return getUserFromHubCredentials()
}

func GetRepoName() (string, error) {
	clusterName, err := status.GetClusterName()
	if err != nil {
		return "", err
	}
	return clusterName + "-wego", nil
}

func getUserFromHubCredentials() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	// check for existing ~/.config/hub
	config, err := ioutil.ReadFile(filepath.Join(homeDir, ".config", "hub"))
	if err != nil {
		return "", err
	}

	data := map[string]interface{}{}
	err = yaml.Unmarshal(config, &data)
	if err != nil {
		return "", err
	}

	return data["github.com"].([]interface{})[0].(map[string]interface{})["user"].(string), nil
}

func isOrganization(owner string) (bool, error) {
	token := os.Getenv("GITHUB_TOKEN")
	response, _, err := utils.CallCommandSeparatingOutputStreams(fmt.Sprintf("curl -u %s:%s https://api.github.com/orgs/%s", owner, token, owner))

	if err != nil {
		return false, err
	}
	var data map[string]interface{}
	err = json.Unmarshal(response, &data)
	if err != nil {
		return false, err
	}
	return data["message"] != "Not Found", nil
}

func initFluxBinary() {
	if fluxBinary == "" {
		fluxPath, err := FluxPath()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to retrieve wego executable path: %v", err)
			os.Exit(1)
		}
		fluxBinary = fluxPath
	}
}
