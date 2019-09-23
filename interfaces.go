package main

import (
  "github.com/concourse/go-concourse/concourse"
  "github.com/concourse/fly/rc"
)

// Interface imports for mocking

type Client interface {
  concourse.Client
}

type Team interface {
  concourse.Team
}

type Target interface {
  rc.Target
}
