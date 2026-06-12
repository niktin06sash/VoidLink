package tun

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

func TestAddAllRoutesRollsBackWhenDefaultRouteFails(t *testing.T) {
	runner := &recordingRunner{
		outputs: map[string][]byte{
			"ip route show default": []byte("default via 192.0.2.1 dev eth0\n"),
		},
		failures: map[string]error{
			"ip route add default via 10.1.1.1 dev void1 metric 500": errors.New("route failed"),
		},
	}
	tun := testTun(runner)

	err := tun.AddAllRoutes("/ip4/203.0.113.10/tcp/12345/p2p/test")
	if err == nil {
		t.Fatal("Expected route setup to fail")
	}

	assertCommands(t, runner.commands,
		"ip route show default",
		"ip route add 203.0.113.10/32 via 192.0.2.1 dev eth0",
		"ip route add default via 10.1.1.1 dev void1 metric 500",
		"ip route del 203.0.113.10/32 via 192.0.2.1 dev eth0",
	)
	if tun.serverIP != "" || tun.dnsConfigured {
		t.Fatal("Failed transaction must not be committed")
	}
}

func TestAddAllRoutesRollsBackRoutesAndPartialDNS(t *testing.T) {
	runner := &recordingRunner{
		outputs: map[string][]byte{
			"ip route show default": []byte("default via 192.0.2.1 dev eth0\n"),
		},
		failures: map[string]error{
			"resolvectl domain void1 ~.": errors.New("domain failed"),
		},
	}
	tun := testTun(runner)

	err := tun.AddAllRoutes("/ip4/203.0.113.10/tcp/12345/p2p/test")
	if err == nil {
		t.Fatal("Expected DNS setup to fail")
	}

	assertCommands(t, runner.commands,
		"ip route show default",
		"ip route add 203.0.113.10/32 via 192.0.2.1 dev eth0",
		"ip route add default via 10.1.1.1 dev void1 metric 500",
		"resolvectl dns void1 8.8.8.8",
		"resolvectl domain void1 ~.",
		"resolvectl revert void1",
		"ip route del default via 10.1.1.1 dev void1",
		"ip route del 203.0.113.10/32 via 192.0.2.1 dev eth0",
	)
	if tun.serverIP != "" || tun.dnsConfigured {
		t.Fatal("Failed transaction must not be committed")
	}
}

func TestAddAndRemoveAllRoutesPreservesSuccessfulBehavior(t *testing.T) {
	runner := &recordingRunner{
		outputs: map[string][]byte{
			"ip route show default": []byte("default via 192.0.2.1 dev eth0\n"),
		},
	}
	tun := testTun(runner)

	if err := tun.AddAllRoutes("/ip4/203.0.113.10/tcp/12345/p2p/test"); err != nil {
		t.Fatal(err)
	}
	tun.RemoveAllRoutes()

	assertCommands(t, runner.commands,
		"ip route show default",
		"ip route add 203.0.113.10/32 via 192.0.2.1 dev eth0",
		"ip route add default via 10.1.1.1 dev void1 metric 500",
		"resolvectl dns void1 8.8.8.8",
		"resolvectl domain void1 ~.",
		"ip route del default via 10.1.1.1 dev void1",
		"ip route del 203.0.113.10/32 via 192.0.2.1 dev eth0",
		"resolvectl revert void1",
	)
	if tun.serverIP != "" || tun.gateway != "" || tun.gwIface != "" || tun.dnsConfigured {
		t.Fatal("Successful cleanup must clear committed route and DNS state")
	}
}

func TestAddSplitRoutesRollsBackWhenDNSFails(t *testing.T) {
	runner := &recordingRunner{
		failures: map[string]error{
			"resolvectl dns void1 8.8.8.8": errors.New("dns failed"),
		},
	}
	tun := testTun(runner)

	err := tun.addSplitRoutes(
		[]string{"198.51.100.0/24"},
		[]string{"203.0.113.0/24", "192.0.2.0/24"},
		"/tmp/routes",
	)
	if err == nil {
		t.Fatal("Expected DNS setup to fail")
	}

	assertCommands(t, runner.commands,
		"ip route add 198.51.100.0/24 via 10.1.1.1 dev void1",
		"ip route add 203.0.113.0/24 via 10.1.1.1 dev void1",
		"ip route add 192.0.2.0/24 via 10.1.1.1 dev void1",
		"resolvectl dns void1 8.8.8.8",
		"ip route del 192.0.2.0/24 via 10.1.1.1 dev void1",
		"ip route del 203.0.113.0/24 via 10.1.1.1 dev void1",
		"ip route del 198.51.100.0/24 via 10.1.1.1 dev void1",
	)
	if len(tun.antifilterRoutes) != 0 || len(tun.defaultRoutes) != 0 || tun.routePath != "" {
		t.Fatal("Failed split route transaction must not be committed")
	}
}

func TestRemoveSplitRoutesRestoresDNSWithoutRoutes(t *testing.T) {
	runner := &recordingRunner{}
	tun := testTun(runner)
	tun.dnsConfigured = true

	tun.RemoveSplitedRoutes()

	assertCommands(t, runner.commands, "resolvectl revert void1")
	if tun.dnsConfigured {
		t.Fatal("DNS state should be cleared after successful restore")
	}
}

func testTun(runner *recordingRunner) *Tun {
	return &Tun{
		ifaceName:  "void1",
		runCommand: runner.run,
	}
}

type recordingRunner struct {
	commands []string
	outputs  map[string][]byte
	failures map[string]error
}

func (r *recordingRunner) run(name string, args ...string) ([]byte, error) {
	command := strings.Join(append([]string{name}, args...), " ")
	r.commands = append(r.commands, command)
	if err := r.failures[command]; err != nil {
		return r.outputs[command], err
	}
	return r.outputs[command], nil
}

func assertCommands(t *testing.T, got []string, expected ...string) {
	t.Helper()
	if !slices.Equal(got, expected) {
		t.Fatalf("Unexpected commands:\n got: %v\nwant: %v", got, expected)
	}
}
