package main

import (
  "github.com/concourse/atc"
  "github.com/concourse/go-concourse/concourse"
	"github.com/ribaptista/concourse-poc/flaghelpers"
  "github.com/concourse/fly/rc"
)

func SetPipeline(target rc.Target, name string, config []byte, vars map[string]string, checkCredentials bool) (bool, bool, []concourse.ConfigWarning, error) {
  _, _, existingConfigVersion, _, err := target.Team().PipelineConfig(name)
	if err != nil {
		if _, ok := err.(concourse.PipelineConfigError); !ok {
			return false, false, nil, err
		}
	}

  newConfig, err := EvaluateConfig(config,
      mapToVarPairs(vars),
      []flaghelpers.YAMLVariablePairFlag{},
      []atc.PathFlag{})
  created, updated, warnings, err := target.Team().CreateOrUpdatePipelineConfig(
		name,
		existingConfigVersion,
		newConfig,
		checkCredentials)
	if err != nil {
		return false, false, nil, err
	}

	return created, updated, warnings, nil
}

func UnpausePipeline(target rc.Target, name string) (bool, error) {
  return target.Team().UnpausePipeline(name)
}

func mapToVarPairs(vars map[string]string) []flaghelpers.VariablePairFlag {
  pairs := []flaghelpers.VariablePairFlag{}
  for k, v := range vars {
    pairs = append(pairs, flaghelpers.VariablePairFlag{
      Name: k,
      Value: v,
    })
  }
  return pairs
}
