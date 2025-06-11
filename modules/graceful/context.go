// Copyright 2019 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package graceful

import (
	"context"
	"errors"
)

var (
	errShutdown  = errors.New("graceful shutdown requested")
	errHammer    = errors.New("graceful hammer requested")
	errTerminate = errors.New("graceful terminate requested")
)

// Shutdown procedure:
// * cancel ShutdownContext: the registered context consumers have time to do their cleanup (they could use the hammer context)
// * cancel HammerContext: the all context consumers have limited time to do their cleanup (wait for a few seconds)
// * cancel TerminateContext: the registered context consumers have time to do their cleanup (but they shouldn't use shutdown/hammer context anymore)
// * cancel manager context
// If the shutdown is triggered again during the shutdown procedure, the hammer context will be canceled immediately to force to shut down.

// ShutdownContext returns a context.Context that is Done at shutdown
// Callers using this context should ensure that they are registered as a running server
// in order that they are waited for.
func (g *Manager) ShutdownContext() context.Context {
	return g.shutdownCtx
}

func (g *Manager) IsShutdownCause(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	return context.Cause(ctx) == errShutdown
}

// HammerContext returns a context.Context that is Done at hammer
// Callers using this context should ensure that they are registered as a running server
// in order that they are waited for.
func (g *Manager) HammerContext() context.Context {
	return g.hammerCtx
}

func (g *Manager) IsHummerCause(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	return context.Cause(ctx) == errHammer
}

// TerminateContext returns a context.Context that is Done at terminate
// Callers using this context should ensure that they are registered as a terminating server
// in order that they are waited for.
func (g *Manager) TerminateContext() context.Context {
	return g.terminateCtx
}

func (g *Manager) IsTerminateCause(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	return context.Cause(ctx) == errTerminate
}

func (g *Manager) IsSystemQuit(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	cause := context.Cause(ctx)
	return cause == errShutdown || cause == errHammer || cause == errTerminate
}
