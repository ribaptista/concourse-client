// Adapted from https://github.com/concourse/fly/blob/66ea6022e466ddcba2b603b1bb40b971e25359fe/commands/internal/setpipelinehelpers/atc_config.go
package main

import (
	"fmt"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"

	"github.com/cloudfoundry/bosh-cli/director/template"
	"github.com/concourse/atc"
  "github.com/ribaptista/concourse-poc/flaghelpers"
	temp "github.com/concourse/fly/template"
)

type BadConfigError struct {
  Warnings []atc.Warning
  Errors []string
}

func (err BadConfigError) Error() string {
  return "Bad configuration"
}

func ValidateConfig(
	configContents []byte,
	templateVariables []flaghelpers.VariablePairFlag,
	yamlTemplateVariables []flaghelpers.YAMLVariablePairFlag,
	templateVariablesFiles []atc.PathFlag,
	strict bool,
) ([]atc.Warning, error) {
	newConfig, err := newConfig(configContents, templateVariablesFiles, templateVariables, yamlTemplateVariables, true, strict)
	if err != nil {
		return nil, err
	}

	var new atc.Config
	if strict {
		// UnmarshalStrict will pick up fields in structs that have the wrong names, as well as any duplicate keys in maps
		// we should consider always using this everywhere in a later release...
		if err := yaml.UnmarshalStrict([]byte(newConfig), &new); err != nil {
			return nil, err
		}
	} else {
		if err := yaml.Unmarshal([]byte(newConfig), &new); err != nil {
			return nil, err
		}
	}

	warnings, errorMessages := new.Validate()

	if len(errorMessages) > 0 || (strict && len(warnings) > 0) {
    return warnings, BadConfigError{
      Warnings: warnings,
      Errors: errorMessages,
    }
	}

	return warnings, nil
}

func EvaluateConfig(configContents []byte,
    templateVariables []flaghelpers.VariablePairFlag,
    yamlTemplateVariables []flaghelpers.YAMLVariablePairFlag,
    templateVariablesFiles []atc.PathFlag) ([]byte, error) {
	newConfig, err := newConfig(configContents, templateVariablesFiles, templateVariables, yamlTemplateVariables, false, false)
	if err != nil {
		return nil, err
	}

	var new atc.Config
	err = yaml.Unmarshal([]byte(newConfig), &new)
	if err != nil {
		return nil, err
	}

	return newConfig, nil
}

func newConfig(
	evaluatedConfig []byte,
	templateVariablesFiles []atc.PathFlag,
	templateVariables []flaghelpers.VariablePairFlag,
	yamlTemplateVariables []flaghelpers.YAMLVariablePairFlag,
	allowEmpty bool,
	strict bool,
) ([]byte, error) {
	if strict {
		// We use a generic map here, since templates are not evaluated yet.
		// (else a template string may cause an error when a struct is expected)
		// If we don't check Strict now, then the subsequent steps will mask any
		// duplicate key errors.
		// We should consider being strict throughout the entire stack by default.
		err := yaml.UnmarshalStrict(evaluatedConfig, make(map[string]interface{}))
		if err != nil {
			return nil, fmt.Errorf("error parsing yaml before applying templates: %s", err.Error())
		}
	}

	var paramPayloads [][]byte
	for _, path := range templateVariablesFiles {
		templateVars, err := ioutil.ReadFile(string(path))
		if err != nil {
			return nil, fmt.Errorf("could not read template variables file (%s): %s", string(path), err.Error())
		}

		paramPayloads = append(paramPayloads, templateVars)
	}

	if temp.Present(evaluatedConfig) {
		resolved, err := resolveDeprecatedTemplateStyle(evaluatedConfig, paramPayloads, templateVariables, yamlTemplateVariables, allowEmpty)
		if err != nil {
			return nil, fmt.Errorf("could not resolve old-style template vars: %s", err.Error())
		}

    evaluatedConfig = resolved
	}

	evaluatedConfig, err := resolveTemplates(evaluatedConfig, paramPayloads, templateVariables, yamlTemplateVariables)
	if err != nil {
		return nil, fmt.Errorf("could not resolve template vars: %s", err.Error())
	}

	return evaluatedConfig, nil
}

func resolveTemplates(configPayload []byte, paramPayloads [][]byte, variables []flaghelpers.VariablePairFlag, yamlVariables []flaghelpers.YAMLVariablePairFlag) ([]byte, error) {
	tpl := template.NewTemplate(configPayload)

	flagVars := template.StaticVariables{}
	for _, f := range variables {
		flagVars[f.Name] = f.Value
	}

	for _, f := range yamlVariables {
		flagVars[f.Name] = f.Value
	}

	vars := []template.Variables{flagVars}
	for i := len(paramPayloads) - 1; i >= 0; i-- {
		payload := paramPayloads[i]

		var staticVars template.StaticVariables
		err := yaml.Unmarshal(payload, &staticVars)
		if err != nil {
			return nil, err
		}

		vars = append(vars, staticVars)
	}

	bytes, err := tpl.Evaluate(template.NewMultiVars(vars), nil, template.EvaluateOpts{})
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func resolveDeprecatedTemplateStyle(
	configPayload []byte,
	paramPayloads [][]byte,
	variables []flaghelpers.VariablePairFlag,
	yamlVariables []flaghelpers.YAMLVariablePairFlag,
	allowEmpty bool,
) ([]byte, error) {
	vars := temp.Variables{}
	for _, payload := range paramPayloads {
		var payloadVars temp.Variables
		err := yaml.Unmarshal(payload, &payloadVars)
		if err != nil {
			return nil, err
		}

		vars = vars.Merge(payloadVars)
	}

	flagVars := temp.Variables{}
	for _, flag := range variables {
		flagVars[flag.Name] = flag.Value
	}

	vars = vars.Merge(flagVars)

	return temp.Evaluate(configPayload, vars, allowEmpty)
}
