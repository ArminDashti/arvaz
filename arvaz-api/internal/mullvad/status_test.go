package mullvad

import "testing"

func TestParseVisibleIPv4(t *testing.T) {
	got := parseVisibleIPv4("Sweden, Stockholm. IPv4: 185.213.154.10, IPv6: 2a03:1b20::1")
	if got != "185.213.154.10" {
		t.Fatalf("got %q", got)
	}
	if parseVisibleIPv4("Sweden, Stockholm") != "" {
		t.Fatal("expected empty")
	}
}

func TestParseTunnelFlags(t *testing.T) {
	raw := `WireGuard options
    MTU:                    unset
    Quantum resistance:     on
    DAITA:                  false
    Public key:             abc
`
	qr, daita := parseTunnelFlags(raw)
	if !qr {
		t.Fatal("expected quantumResistant true")
	}
	if daita {
		t.Fatal("expected daita false")
	}

	raw2 := "Quantum resistance: off\nDAITA: on\n"
	qr, daita = parseTunnelFlags(raw2)
	if qr || !daita {
		t.Fatalf("got qr=%v daita=%v", qr, daita)
	}
}
