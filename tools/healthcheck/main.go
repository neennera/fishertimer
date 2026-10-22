// healthcheck is a tiny HTTP probe for Docker HEALTHCHECKs.
//
// The service images are distroless (no shell, no curl or wget), so a
// container cannot probe itself with the usual tools. This static binary is
// copied into every Go image at /healthcheck:
//
//	/healthcheck http://127.0.0.1:8082/health
//
// It exits 0 when the URL answers with a 2xx status, 1 otherwise.
package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const timeout = 3 * time.Second

// probe returns nil when url answers with a 2xx status within the timeout.
func probe(url string) error {
	client := http.Client{Timeout: timeout}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("%s answered %s", url, resp.Status)
	}
	return nil
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: healthcheck <url>")
		os.Exit(2)
	}
	if err := probe(os.Args[1]); err != nil {
		fmt.Fprintln(os.Stderr, "unhealthy:", err)
		os.Exit(1)
	}
}
