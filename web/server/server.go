//-----------------------------------------------------------------------------
// Copyright (c) 2020 Detlef Stern
//
// This file is part of zettelstore.
//
// Zettelstore is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

// Package server provides a web server.
package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Server timeout values
const (
	shutdownTimeout = 5 * time.Second
	readTimeout     = 5 * time.Second
	writeTimeout    = 10 * time.Second
	idleTimeout     = 120 * time.Second
)

// Start a web server.
func Start(addr string, handler http.Handler) error {
	srv := newServer(addr, handler)
	waitInterrupt := make(chan os.Signal)
	waitError := make(chan error)
	signal.Notify(waitInterrupt, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-waitInterrupt
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			waitError <- err
			return
		}
		waitError <- nil
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return <-waitError
}

type server struct {
	*http.Server
}

func newServer(addr string, handler http.Handler) *server {
	if addr == "" {
		addr = ":http"
	}
	srv := &server{
		Server: &http.Server{
			Addr:    addr,
			Handler: handler,

			// See: https://blog.cloudflare.com/exposing-go-on-the-internet/
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
			IdleTimeout:  idleTimeout,
		},
	}
	return srv
}

func (srv *server) Shutdown(ctx context.Context) error {
	// TODO: prepare shutdown
	err := srv.Server.Shutdown(ctx)
	// TODO: after shutdown(err)
	return err
}
