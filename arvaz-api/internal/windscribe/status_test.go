package windscribe

import "testing"

func TestParseStatus(t *testing.T) {
	raw := `Internet connectivity: Available
Login state: Logged in
Firewall state: Off
Connect state: Connected: Manchester - City
Protocol: Stealth:443
IP: 84.233.178.40
`
	st := parseStatus(raw)
	if !st.Connected {
		t.Fatal("expected connected")
	}
	if st.Location != "Manchester - City" {
		t.Fatalf("location %q", st.Location)
	}
	if st.PublicIP != "84.233.178.40" {
		t.Fatalf("ip %q", st.PublicIP)
	}
	if st.Country != "Manchester" || st.City != "City" {
		t.Fatalf("country/city %q / %q", st.Country, st.City)
	}
}

func TestParseLocationList(t *testing.T) {
	raw := `* UK - Manchester - City
US - Atlanta - Piedmont
`
	locs := parseLocationList(raw, "Manchester - City")
	if len(locs) != 2 {
		t.Fatalf("got %d locations", len(locs))
	}
	if !locs[0].Active {
		t.Fatal("expected first active")
	}
	if locs[1].Region != "US" || locs[1].City != "Atlanta" {
		t.Fatalf("second loc %+v", locs[1])
	}
}
