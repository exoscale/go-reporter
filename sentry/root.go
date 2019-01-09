// Package sentry integrates Sentry with NIDWALD.
//
// This is really a thin wrapper around Sentry. Not much is happening
// here.
package sentry

import (
	"github.com/getsentry/raven-go"
	"github.com/pkg/errors"
)

func init() {
	raven.Close()
}

// New creates a new sentry client from configuration. If no DSN is
// provided, return nil.
func New(config Configuration) (*raven.Client, error) {
	if config.DSN == "" {
		return nil, nil
	}
	client, err := raven.NewWithTags(config.DSN, config.Tags)
	if err != nil {
		return nil, errors.Wrap(err, "unable to initialize Sentry client")
	}
	if config.Version != "" {
		client.SetRelease(config.Version)
	}
	return client, nil
}
