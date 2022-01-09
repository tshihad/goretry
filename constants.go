package goretry

import "time"

const (
	defaultRetryDelay = time.Second * 5
	defaultRetryCount = 3
	defaultTimeout    = time.Hour * 24
	noRetryCount      = -1
)
