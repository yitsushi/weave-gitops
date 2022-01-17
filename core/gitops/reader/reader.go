package reader

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"

	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"
)

func readJsonOrYamlFile(fileSystem fs.FS, path string) (map[string]interface{}, error) {
	r, err := fileSystem.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open file %s: %w", path, err)
	}

	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	jsonData, err := yaml.YAMLToJSON(data)
	if err != nil {
		return nil, fmt.Errorf("an error occurred converting to yaml: %w", err)
	}

	var obj map[string]interface{}
	err = json.Unmarshal(jsonData, &obj)
	if err != nil {
		return nil, fmt.Errorf("an error occurred unmarshalling to json: %w", err)
	}

	return obj, nil
}

func readKustomizationFile(fileSystem fs.FS, path string) (types.Kustomization, error) {
	r, err := fileSystem.Open(path)
	if err != nil {
		return types.Kustomization{}, fmt.Errorf("cannot open file %s: %w", path, err)
	}

	data, err := ioutil.ReadAll(r)
	if err != nil {
		return types.Kustomization{}, fmt.Errorf("error reading file: %w", err)
	}

	var kust types.Kustomization
	err = yaml.Unmarshal(data, &kust)
	if err != nil {
		return types.Kustomization{}, fmt.Errorf("error unmarshaling k8s kustomization file: %w", err)
	}

	return kust, nil
}
