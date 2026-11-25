package agscheduler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// WaitForShutdown blocks until SIGINT or SIGTERM is received. It calls
// beforeShutdown, then gracefully shuts down the scheduler within timeout.
// A second signal terminates the process immediately with exit code 1.
func (s *Scheduler) WaitForShutdown(timeout time.Duration, beforeShutdown func() error) error {
	signals := make(chan os.Signal, 2)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signals)

	return s.waitForShutdown(timeout, beforeShutdown, signals, func() {
		slog.Warn("Second shutdown signal received, forcing exit.")
		os.Exit(1)
	})
}

func (s *Scheduler) waitForShutdown(
	timeout time.Duration,
	beforeShutdown func() error,
	signals <-chan os.Signal,
	forceExit func(),
) error {
	sig, ok := <-signals
	if !ok {
		return nil
	}
	slog.Info(fmt.Sprintf("Shutdown signal `%s` received, waiting for running jobs.", sig))

	shutdownDone := make(chan struct{})
	defer close(shutdownDone)
	go func() {
		select {
		case <-signals:
			forceExit()
		case <-shutdownDone:
		}
	}()

	var errs []error
	if beforeShutdown != nil {
		if err := beforeShutdown(); err != nil {
			errs = append(errs, fmt.Errorf("shutdown services: %w", err))
		}
	}

	if timeout <= 0 {
		timeout = DefaultShutdownTimeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}
