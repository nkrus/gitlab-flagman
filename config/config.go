package config

import (
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// FeatureFlag структура для чтения флагов из файла
type FeatureFlag struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description" json:"description"`
	Active      bool   `yaml:"active" json:"active"`
	Strategies  []struct {
		Name       string                 `yaml:"name" json:"name"`
		Parameters map[string]interface{} `yaml:"parameters" json:"parameters"`
		Scopes     []struct {
			Environment string `yaml:"environment_scope" json:"environment_scope"`
		} `yaml:"scopes" json:"scopes"`
	} `yaml:"strategies" json:"strategies"`
}

func ReadFlagsFromYAML(fileName string) ([]FeatureFlag, error) {

	if !strings.HasSuffix(fileName, ".yaml") {
		return nil, fmt.Errorf("flags file must have .yaml extension")
	}

	// Open the file
	yamlFile, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer yamlFile.Close()

	// Read the file contents
	fileContent, err := io.ReadAll(yamlFile)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	// Parse the YAML
	var featureFlags []FeatureFlag
	if err := yaml.Unmarshal(fileContent, &featureFlags); err != nil {
		return nil, fmt.Errorf("error unmarshalling YAML: %w", err)
	}

	return featureFlags, nil
}
