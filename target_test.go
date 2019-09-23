package main

import (
  "testing"
  "github.com/concourse/atc"
  "github.com/stretchr/testify/assert"
  "github.com/stretchr/testify/mock"
  "github.com/ribaptista/concourse-poc/mocks"
  "github.com/concourse/fly/rc"
)

func TestMatchingVersion(t *testing.T) {
  token := &rc.TargetToken{
    Type:  "bearer",
    Value: "bar",
  }
  authenticator := new(mocks.Authenticator)
  authenticator.On("GetToken", mock.Anything,
      mock.Anything, mock.Anything).Return(token, nil)
  client := new(mocks.Client)
  clientFactory := new(mocks.ClientFactory)
  clientFactory.On("NewClient", mock.Anything,
        mock.Anything, mock.Anything).Return(client)
  client.On("GetInfo").Return(atc.Info{
    Version: "4.2.5",
    WorkerVersion: "0.0.0",
  }, nil)
  target, err := NewAuthenticatedTarget(
    "foo",
    "http://concourse.localhost/concourse",
    "user",
    "pass",
    "coolteam",
    "",
    true,
    false,
    clientFactory,
    authenticator)
  err = target.Validate()
  assert.Nil(t, err, "Should match versions")
}

func TestUnmatchingVersion(t *testing.T) {
  token := &rc.TargetToken{
    Type:  "bearer",
    Value: "bar",
  }
  authenticator := new(mocks.Authenticator)
  authenticator.On("GetToken", mock.Anything,
      mock.Anything, mock.Anything).Return(token, nil)
  client := new(mocks.Client)
  clientFactory := new(mocks.ClientFactory)
  clientFactory.On("NewClient", mock.Anything,
        mock.Anything, mock.Anything).Return(client)
  client.On("GetInfo").Return(atc.Info{
    Version: "5.2.5",
    WorkerVersion: "0.0.0",
  }, nil)
  target, err := NewAuthenticatedTarget(
    "foo",
    "http://concourse.localhost/concourse",
    "user",
    "pass",
    "coolteam",
    "",
    true,
    false,
    clientFactory,
    authenticator)
  err = target.Validate()
  assert.IsType(t, err, ErrVersionMismatch{}, "Should detect unmatching versions")
}
