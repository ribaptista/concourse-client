package main

import (
  "testing"
  "github.com/stretchr/testify/assert"
  "github.com/ribaptista/concourse-poc/flaghelpers"
  "github.com/concourse/atc"
  yaml "gopkg.in/yaml.v2"
)

func TestValidateEmpty(t *testing.T) {
  _, err := ValidateConfig([]byte("jobs: []"),
    []flaghelpers.VariablePairFlag{},
    []flaghelpers.YAMLVariablePairFlag{},
    []atc.PathFlag{},
    false)
  assert.Nil(t, err, "Should parse empty config")
}

func TestValidateBadYaml(t *testing.T) {
  _, err := ValidateConfig([]byte("invalid"),
    []flaghelpers.VariablePairFlag{},
    []flaghelpers.YAMLVariablePairFlag{},
    []atc.PathFlag{},
    false)
  assert.IsType(t, err, &yaml.TypeError{}, "Should fail to parse bad yaml")
}

func TestValidateDuplicateJobs(t *testing.T) {
  _, err := ValidateConfig([]byte(`
jobs:
  - name: foo
  - name: foo`),
    []flaghelpers.VariablePairFlag{},
    []flaghelpers.YAMLVariablePairFlag{},
    []atc.PathFlag{},
    false)
  assert.IsType(t, err, BadConfigError{}, "Should detect duplicate jobs")
}

func TestEvaluateEmptyPipeline(t *testing.T) {
  _, err := EvaluateConfig([]byte(``),
    []flaghelpers.VariablePairFlag{},
    []flaghelpers.YAMLVariablePairFlag{},
    []atc.PathFlag{})
  assert.Nil(t, err, "Should evaluate empty pipeline")
}

func TestParseVars(t *testing.T) {
  configYaml, err := EvaluateConfig([]byte(`jobs:
  - name: ((jobName))`),
    []flaghelpers.VariablePairFlag{
      flaghelpers.VariablePairFlag{
        Name: "jobName",
        Value: "importantJob",
      },
    },
    []flaghelpers.YAMLVariablePairFlag{},
    []atc.PathFlag{})
  var config atc.Config
  _ = yaml.Unmarshal(configYaml, &config)
  assert.Nil(t, err, "Should set empty pipeline")
  assert.Equal(t, config.Jobs[0].Name, "importantJob", "Should parse vars")
}

func TestParseOldStyleVars(t *testing.T) {
  configYaml, err := EvaluateConfig([]byte(`jobs:
  - name: {{jobName}}`),
    []flaghelpers.VariablePairFlag{
      flaghelpers.VariablePairFlag{
        Name: "jobName",
        Value: "importantJob",
      },
    },
    []flaghelpers.YAMLVariablePairFlag{},
    []atc.PathFlag{})
  var config atc.Config
  _ = yaml.Unmarshal(configYaml, &config)
  assert.Nil(t, err, "Should set empty pipeline")
  assert.Equal(t, config.Jobs[0].Name, "importantJob", "Should parse old style vars")
}
