package main

import (
  "errors"
  "testing"
  "github.com/concourse/atc"
  "github.com/stretchr/testify/assert"
  "github.com/stretchr/testify/mock"
  "github.com/ribaptista/concourse-poc/mocks"
  "github.com/concourse/go-concourse/concourse"
  yaml "gopkg.in/yaml.v2"
)

func TestSetEmptyPipeline(t *testing.T) {
  team := new(mocks.Team)
  team.On("PipelineConfig", "foo").Return(
      atc.Config{},
      atc.RawConfig(""),
      "",
      false,
      nil)
  team.On("CreateOrUpdatePipelineConfig", "foo", "", mock.Anything, false).Return(
      true,
      false,
      []concourse.ConfigWarning{},
      nil)
  target := new(mocks.Target)
  target.On("Team").Return(team)
  _, _, _, err := SetPipeline(target,
    "foo",
    []byte(``),
    map[string]string{},
    false)
  assert.Nil(t, err, "Should create empty pipeline")
}

func TestSetTemplatedPipeline(t *testing.T) {
  team := new(mocks.Team)
  team.On("PipelineConfig", "foo").Return(
      atc.Config{},
      atc.RawConfig(""),
      "",
      false,
      nil)
  team.On("CreateOrUpdatePipelineConfig", "foo", "", mock.MatchedBy(func (configYaml []byte) bool {
      var config atc.Config
      err := yaml.Unmarshal(configYaml, &config)
      return err == nil && config.Jobs[0].Name == "importantJob"
    }), false).Return(
      true,
      false,
      []concourse.ConfigWarning{},
      nil)
  target := new(mocks.Target)
  target.On("Team").Return(team)
  _, _, _, err := SetPipeline(target,
    "foo",
    []byte(`jobs: [ name: ((jobName)) ]`),
    map[string]string{
      "jobName": "importantJob",
    },
    false)
  assert.Nil(t, err, "Should parse template variables")
}

func TestConcourseFailure(t *testing.T) {
  team := new(mocks.Team)
  team.On("PipelineConfig", "foo").Return(
      atc.Config{},
      atc.RawConfig(""),
      "",
      false,
      nil)
  team.On("CreateOrUpdatePipelineConfig", "foo", "", mock.Anything, false).Return(
      false,
      false,
      []concourse.ConfigWarning{},
      errors.New("Server failure"))
  target := new(mocks.Target)
  target.On("Team").Return(team)
  _, _, _, err := SetPipeline(target,
    "foo",
    []byte(``),
    map[string]string{},
    false)
  assert.NotNil(t, err, "Should receive error from concourse server")
}
