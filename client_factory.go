package main

import (
  "net/http"
  "github.com/concourse/go-concourse/concourse"
)

type ClientFactory interface {
  NewClient(url string, httpClient *http.Client, tracing bool) concourse.Client
}

type ConcourseClientFactory struct {}

func (f *ConcourseClientFactory) NewClient(url string, httpClient *http.Client, tracing bool) concourse.Client {
  return concourse.NewClient(url, httpClient, tracing)
}
