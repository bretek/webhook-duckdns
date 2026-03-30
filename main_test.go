package main

import (
	"os"
	"testing"

	acmetest "github.com/cert-manager/cert-manager/test/acme"
)

var (
	zone = os.Getenv("TEST_ZONE_NAME")
)

func TestRunsSuite(t *testing.T) {
	fixture := acmetest.NewFixture(&duckDNSProviderSolver{},
		acmetest.SetResolvedZone(zone),
		acmetest.SetManifestPath("testdata/duckdns"),
		acmetest.SetAllowAmbientCredentials(false),
		acmetest.SetDNSServer("ns1.duckdns.org:53"),
	)

	fixture.RunConformance(t)
}
