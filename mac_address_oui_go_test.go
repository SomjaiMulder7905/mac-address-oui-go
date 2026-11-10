package macaddressouigo

import (
	"strings"
	"testing"
)

func TestDefaultLookupKnown(t *testing.T) {
	r := Default()
	cases := []struct {
		mac  string
		want string
	}{
		{"00:00:0C:01:02:03", "Cisco Systems, Inc"},
		{"00-00-5E-AA-BB-CC", "ICANN, IANA Department"},
		{"0002B3DEADBE", "Intel Corporation"},
		{"001b.63aa.bbcc", "Apple, Inc."},
		{"FCFBFB000001", "Cisco Systems, Inc"},
	}
	for _, tc := range cases {
		got, ok := r.Lookup(tc.mac)
		if !ok {
			t.Fatalf("Lookup(%q) returned not-found", tc.mac)
		}
		if got != tc.want {
			t.Errorf("Lookup(%q) = %q, want %q", tc.mac, got, tc.want)
		}
	}
}

func TestLookupUnknown(t *testing.T) {
	r := Default()
	if _, ok := r.Lookup("AA:BB:CC:DD:EE:FF"); ok {
		t.Fatal("expected unknown OUI to miss")
	}
}

func TestLookupMalformed(t *testing.T) {
	r := Default()
	bad := []string{
		"",
		"not-a-mac",
		"00:00:0C:01:02",       // too short
		"00:00:0C:01:02:03:04", // too long
		"00:00:0G:01:02:03",    // non-hex
	}
	for _, mac := range bad {
		if _, ok := r.Lookup(mac); ok {
			t.Errorf("Lookup(%q) unexpectedly succeeded", mac)
		}
	}
}

func TestAddNormalizesSeparators(t *testing.T) {
	r := NewResolver()
	if err := r.Add("aa-bb-cc", "Example Org"); err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	got, ok := r.Lookup("AA:BB:CC:00:00:00")
	if !ok || got != "Example Org" {
		t.Fatalf("Lookup after Add = (%q, %v), want (\"Example Org\", true)", got, ok)
	}
}

func TestAddRejectsInvalidOUI(t *testing.T) {
	r := NewResolver()
	bad := []string{"", "12345", "1234567", "ZZZZZZ", "12:34:5"}
	for _, oui := range bad {
		if err := r.Add(oui, "x"); err == nil {
			t.Errorf("Add(%q) unexpectedly succeeded", oui)
		}
	}
}

func TestAddOverwrites(t *testing.T) {
	r := NewResolver()
	if err := r.Add("00000C", "First"); err != nil {
		t.Fatal(err)
	}
	if err := r.Add("00000C", "Second"); err != nil {
		t.Fatal(err)
	}
	got, ok := r.Lookup("00:00:0C:00:00:00")
	if !ok || got != "Second" {
		t.Fatalf("got (%q, %v), want (\"Second\", true)", got, ok)
	}
}

func TestNewResolverFromCSV(t *testing.T) {
	data := `# comment line

00:00:0C,Cisco Systems, Inc
00-00-5E,ICANN, IANA Department
0002B3,"Intel, Corporation"
`
	r, err := NewResolverFromCSV(strings.NewReader(data))
	if err != nil {
		t.Fatalf("NewResolverFromCSV: %v", err)
	}
	if r.Len() != 3 {
		t.Fatalf("Len() = %d, want 3", r.Len())
	}
	got, ok := r.Lookup("00:02:B3:11:22:33")
	if !ok || got != "Intel, Corporation" {
		t.Fatalf("Lookup = (%q, %v), want (\"Intel, Corporation\", true)", got, ok)
	}
}

func TestNewResolverFromCSVErrors(t *testing.T) {
	cases := []string{
		"00:00:0C\n",          // only one field
		"ZZZZZZ,Org\n",        // invalid OUI
		"00:00:0C,\"unclosed\n", // bad quoting
	}
	for _, data := range cases {
		if _, err := NewResolverFromCSV(strings.NewReader(data)); err == nil {
			t.Errorf("NewResolverFromCSV(%q) unexpectedly succeeded", data)
		}
	}
}

func TestDefaultLen(t *testing.T) {
	if got := Default().Len(); got != 14 {
		t.Fatalf("Default().Len() = %d, want 14", got)
	}
}
