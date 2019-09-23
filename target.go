// Adapted from https://github.com/concourse/fly/blob/66ea6022e466ddcba2b603b1bb40b971e25359fe/rc/targets.go
package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
	"runtime"
	"time"

	"github.com/concourse/atc"
	"github.com/concourse/fly/version"
	"github.com/concourse/go-concourse/concourse"
	semisemanticversion "github.com/cppforlife/go-semi-semantic/version"
  "github.com/concourse/fly/rc"
	"golang.org/x/oauth2"
)

const (
	FlyVersion = "4.2.5"
)

type ErrVersionMismatch struct {
	flyVersion string
	atcVersion string
	targetName rc.TargetName
}

func NewErrVersionMismatch(flyVersion string, atcVersion string, targetName rc.TargetName) ErrVersionMismatch {
	return ErrVersionMismatch{
		flyVersion: flyVersion,
		atcVersion: atcVersion,
		targetName: targetName,
	}
}

func (e ErrVersionMismatch) Error() string {
	return fmt.Sprintf("Version mismatch: client is %s, server is %s", e.flyVersion, e.atcVersion)
}

type target struct {
	name      rc.TargetName
	teamName  string
	caCert    string
	tlsConfig *tls.Config
	client    concourse.Client
	url       string
	token     *rc.TargetToken
	info      atc.Info
}

func newTarget(
	name rc.TargetName,
	teamName string,
	url string,
	token *rc.TargetToken,
	caCert string,
	caCertPool *x509.CertPool,
	insecure bool,
	client concourse.Client,
) *target {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: insecure,
		RootCAs:            caCertPool,
	}

	return &target{
		name:      name,
		teamName:  teamName,
		url:       url,
		token:     token,
		caCert:    caCert,
		tlsConfig: tlsConfig,
		client:    client,
	}
}

func NewAuthenticatedTarget(
	name rc.TargetName,
	url string,
  username string,
	password string,
	teamName string,
	caCert string,
	insecure bool,
	tracing bool,
	clientFactory ClientFactory,
	authenticator Authenticator,
) (rc.Target, error) {
	caCertPool, err := loadCACertPool(caCert)
	if err != nil {
		return nil, err
	}

	token, err := authenticate(url, username, password, caCertPool,
			insecure, tracing, clientFactory, authenticator)
	if err != nil {
		return nil, err
	}

  httpClient := defaultHttpClient(token, insecure, caCertPool)
	client := clientFactory.NewClient(url, httpClient, tracing)
	return newTarget(
		name,
		teamName,
		url,
		token,
		caCert,
		caCertPool,
		insecure,
		client,
	), nil
}

func authenticate(
		url string,
		username string,
		password string,
		caCertPool *x509.CertPool,
		insecure bool,
		tracing bool,
		clientFactory ClientFactory,
		authenticator Authenticator) (*rc.TargetToken, error) {
	httpClient := &http.Client{Transport: transport(insecure, caCertPool)}
	client := clientFactory.NewClient(url, httpClient, tracing)
	token, err := authenticator.GetToken(client, username, password)
  if err != nil {
    return nil, errors.New(fmt.Sprintf("Failed to authenticate: %s", err.Error()))
  }

	return token, nil
}

func (t *target) Client() concourse.Client {
	return t.client
}

func (t *target) Team() concourse.Team {
	return t.client.Team(t.teamName)
}

func (t *target) CACert() string {
	return t.caCert
}

func (t *target) TLSConfig() *tls.Config {
	return t.tlsConfig
}

func (t *target) URL() string {
	return t.url
}

func (t *target) Token() *rc.TargetToken {
	return t.token
}

func (t *target) Version() (string, error) {
	info, err := t.getInfo()
	if err != nil {
		return "", err
	}

	return info.Version, nil
}

func (t *target) WorkerVersion() (string, error) {
	info, err := t.getInfo()
	if err != nil {
		return "", err
	}

	return info.WorkerVersion, nil
}

func (t *target) TokenAuthorization() (string, bool) {
	if t.token == nil || (t.token.Type == "" && t.token.Value == "") {
		return "", false
	}

	return t.token.Type + " " + t.token.Value, true
}

func (t *target) ValidateWithWarningOnly() error {
	return nil
}

func (t *target) Validate() error {
	return t.validate()
}

func (t *target) IsWorkerVersionCompatible(workerVersion string) (bool, error) {
	info, err := t.getInfo()
	if err != nil {
		return false, err
	}

	if info.WorkerVersion == "" {
		return true, nil
	}

	if workerVersion == "" {
		return false, nil
	}

	workerV, err := semisemanticversion.NewVersionFromString(workerVersion)
	if err != nil {
		return false, err
	}

	infoV, err := semisemanticversion.NewVersionFromString(info.WorkerVersion)
	if err != nil {
		return false, err
	}

	if workerV.Release.Components[0].Compare(infoV.Release.Components[0]) != 0 {
		return false, nil
	}

	if workerV.Release.Components[1].Compare(infoV.Release.Components[1]) == -1 {
		return false, nil
	}

	return true, nil
}

func (t *target) validate() error {
	info, err := t.getInfo()
	if err != nil {
		return err
	}

	if info.Version == FlyVersion {
		return nil
	}

	atcMajor, atcMinor, _, err := version.GetSemver(info.Version)
	if err != nil {
		return err
	}

	flyMajor, flyMinor, _, err := version.GetSemver(FlyVersion)
	if err != nil {
		return err
	}

	if atcMajor != flyMajor || atcMinor != flyMinor {
		return NewErrVersionMismatch(FlyVersion, info.Version, t.name)
	}

	return nil
}

func (t *target) getInfo() (atc.Info, error) {
	if (t.info != atc.Info{}) {
		return t.info, nil
	}

	var err error
	t.info, err = t.client.GetInfo()
	return t.info, err
}

func defaultHttpClient(token *rc.TargetToken, insecure bool, caCertPool *x509.CertPool) *http.Client {
	var oAuthToken *oauth2.Token
	if token != nil {
		oAuthToken = &oauth2.Token{
			TokenType:   token.Type,
			AccessToken: token.Value,
		}
	}

	transport := transport(insecure, caCertPool)

	if token != nil {
		transport = &oauth2.Transport{
			Source: oauth2.StaticTokenSource(oAuthToken),
			Base:   transport,
		}
	}

	return &http.Client{Transport: transport}
}

func loadCACertPool(caCert string) (cert *x509.CertPool, err error) {
	if caCert == "" {
		return nil, nil
	}

	// TODO: remove else block once we switch to go 1.8
	// x509.SystemCertPool is not supported in go 1.7 on Windows
	// see: https://github.com/golang/go/issues/16736
	var pool *x509.CertPool
	if runtime.GOOS != "windows" {
		var err error
		pool, err = x509.SystemCertPool()
		if err != nil {
			return nil, err
		}
	} else {
		pool = x509.NewCertPool()
	}

	ok := pool.AppendCertsFromPEM([]byte(caCert))
	if !ok {
		return nil, errors.New("CA Cert not valid")
	}
	return pool, nil
}

func transport(insecure bool, caCertPool *x509.CertPool) http.RoundTripper {
	var transport http.RoundTripper

	transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecure,
			RootCAs:            caCertPool,
		},
		Dial: (&net.Dialer{
			Timeout: 10 * time.Second,
		}).Dial,
		Proxy: http.ProxyFromEnvironment,
	}

	return transport
}
