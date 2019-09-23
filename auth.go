package main

import (
  "context"
  "github.com/concourse/go-concourse/concourse"
  "github.com/concourse/fly/rc"
  "golang.org/x/oauth2"
)

type Authenticator interface {
  GetToken(client concourse.Client, username, password string) (*rc.TargetToken, error)
}

type OAuth2Authenticator struct {}

func (a *OAuth2Authenticator) GetToken(client concourse.Client, username, password string) (*rc.TargetToken, error) {
  // TODO: Check OIDC and Octa
  oauth2Config := oauth2.Config{
    ClientID:     "fly",
    ClientSecret: "Zmx5",
    Endpoint:     oauth2.Endpoint{TokenURL: client.URL() + "/sky/token"},
    Scopes:       []string{"openid", "profile", "email", "federated:id", "groups"},
  }
  ctx := context.WithValue(context.Background(), oauth2.HTTPClient, client.HTTPClient())
  token, err := oauth2Config.PasswordCredentialsToken(ctx, username, password)
  if err != nil {
    return nil, err
  }

  return &rc.TargetToken{
    Type:  token.TokenType,
    Value: token.AccessToken,
  }, nil}
