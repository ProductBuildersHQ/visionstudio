package main

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

// dsnAddr extracts the host:port from a MySQL DSN of the form
// "user:pass@tcp(host:port)/db". It returns "the configured address" when the
// DSN cannot be parsed, so diagnostics never print an empty target.
func dsnAddr(dsn string) string {
	_, rest, ok := strings.Cut(dsn, "tcp(")
	if !ok {
		return "the configured address"
	}
	addr, _, ok := strings.Cut(rest, ")")
	if !ok {
		return "the configured address"
	}
	return addr
}

// portFromDSN extracts the numeric port from a MySQL DSN's host:port segment.
// It returns 0 when the port cannot be determined.
func portFromDSN(dsn string) int {
	_, portStr, ok := strings.Cut(dsnAddr(dsn), ":")
	if !ok {
		return 0
	}
	p, err := strconv.Atoi(portStr)
	if err != nil {
		return 0
	}
	return p
}

// diagnoseDBError inspects an error from connecting to (or pinging) the Dolt
// server and, when it looks like a lifecycle/setup problem, rewrites it into an
// actionable message that tells the caller how to fix it. This is the message
// other agents see when they reach visionstudio expecting a live database.
// Errors that are not recognized connectivity problems are returned unchanged.
func diagnoseDBError(dsn string, err error) error {
	if err == nil {
		return nil
	}
	addr := dsnAddr(dsn)
	msg := err.Error()

	// Server not running / host unreachable — the common case when the local
	// Dolt server has not been started.
	var netErr net.Error
	if strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "dial tcp") ||
		strings.Contains(msg, "no such host") ||
		strings.Contains(msg, "i/o timeout") ||
		(errors.As(err, &netErr) && netErr.Timeout()) {
		return fmt.Errorf(`cannot reach the VisionStudio database at %s — the Dolt SQL server is not running.

Start it, then retry your command:

    visionstudio db start      # start the Dolt SQL server in the background
    visionstudio db status     # verify it is running

If the server runs on a different port, point visionstudio at it (this is also saved
to ~/.productbuildershq/visionstudio/config.json by db start/serve), e.g.:

    export VISIONSTUDIO_DSN='root:@tcp(127.0.0.1:PORT)/visionstudio'

(underlying error: %w)`, addr, err)
	}

	// Server reachable, but the database/schema is missing.
	if strings.Contains(msg, "Unknown database") || strings.Contains(msg, "Error 1049") {
		return fmt.Errorf(`connected to the Dolt server at %s, but the "visionstudio" database has not been initialized.

Initialize and migrate it, then retry:

    visionstudio db init --migrate

(underlying error: %w)`, addr, err)
	}

	// Server reachable, but credentials are rejected.
	if strings.Contains(msg, "Access denied") || strings.Contains(msg, "Error 1045") {
		return fmt.Errorf(`the Dolt server at %s rejected the configured credentials.

Check your DSN (--dsn flag, $VISIONSTUDIO_DSN, or config file). Current target: %s

(underlying error: %w)`, addr, dsn, err)
	}

	return err
}
