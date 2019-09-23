package main

import (
  "log"
  "io/ioutil"
)

const (
  targetName = "ricardo"
  atcUrl = "http://concourse:8080"
  team = "main"
  insecure = true
  caCert = ""
  tracing = true
  username = "test"
  password = "test"
  pipelineName = "hello-world"
  checkCredentials = false // ?
  configPath = "./pipeline.yml"
)

var vars = map[string]string{
  "name": "c6",
}

func main() {
  target, err := NewAuthenticatedTarget(targetName, atcUrl, username, password,
      team, caCert, insecure, tracing, &ConcourseClientFactory{},
      &OAuth2Authenticator{});
  if err != nil {
    log.Fatalf("Could not init target: %s", err.Error())
  }

  evaluatedConfig, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Could not read config: %s", err.Error())
	}

  _, _, _, err = SetPipeline(target, pipelineName, evaluatedConfig, vars, false)
  if err != nil {
    log.Fatalf("Could not set pipeline: %s", err.Error())
  }

  _, err = UnpausePipeline(target, pipelineName);
  if err != nil {
		log.Fatalf("Failed to unpause pipeline %s", err.Error())
	}
}
