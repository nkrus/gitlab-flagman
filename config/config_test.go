package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadFlagsFromYAML(t *testing.T) {
	t.Run("successfully read and parse valid YAML", func(t *testing.T) {
		yamlContent := `
- name: "Feature1"
  description: "First feature flag"
  active: true
  strategies:
    - name: "default"
      parameters: {}
      scopes:
        - environment_scope: "TEST"
`
		tmpFile, err := os.CreateTemp("", "feature_flags_*.yaml")
		assert.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		_, err = tmpFile.WriteString(yamlContent)
		assert.NoError(t, err)
		tmpFile.Close()

		flags, err := ReadFlagsFromYAML(tmpFile.Name())

		assert.NoError(t, err)
		assert.Len(t, flags, 1)
		assert.Equal(t, "Feature1", flags[0].Name)
		assert.Equal(t, "First feature flag", flags[0].Description)
		assert.True(t, flags[0].Active)
		assert.Len(t, flags[0].Strategies, 1)
		assert.Equal(t, "default", flags[0].Strategies[0].Name)
		assert.Equal(t, "TEST", flags[0].Strategies[0].Scopes[0].Environment)
	})

	t.Run("file with invalid extension", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "feature_flags_*.txt")
		assert.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		_, err = tmpFile.WriteString("not a valid yaml content")
		assert.NoError(t, err)
		tmpFile.Close()

		flags, err := ReadFlagsFromYAML(tmpFile.Name())

		assert.Error(t, err)
		assert.Nil(t, flags)
		assert.Equal(t, "flags file must have .yaml extension", err.Error())
	})

	t.Run("file not found", func(t *testing.T) {
		flags, err := ReadFlagsFromYAML("non_existent_file.yaml")

		assert.Error(t, err)
		assert.Nil(t, flags)
		assert.Contains(t, err.Error(), "error opening file")
	})

	t.Run("error unmarshalling invalid YAML", func(t *testing.T) {
		invalidYAML := `
- name: "Feature1"
  description: "First feature flag"
  active: true
  strategies:
    - name: "default"
      parameters: 
        - invalid_list
`

		tmpFile, err := os.CreateTemp("", "feature_flags_*.yaml")
		assert.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		_, err = tmpFile.WriteString(invalidYAML)
		assert.NoError(t, err)
		tmpFile.Close()

		flags, err := ReadFlagsFromYAML(tmpFile.Name())

		assert.Error(t, err)
		assert.Nil(t, flags)
		assert.Contains(t, err.Error(), "error unmarshalling YAML")
	})
}
