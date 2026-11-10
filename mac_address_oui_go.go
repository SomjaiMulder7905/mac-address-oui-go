// Package macaddressouigo resolves the Organizationally Unique Identifier (OUI)
// portion of a MAC address against an in-memory table of IEEE assignments.
//
// The package deliberately embeds a small, curated subset of the IEEE registry
// rather than downloading the full public listing. That keeps the library
// dependency-free and usable in offline environments. Callers that need the
// full registry can load it with NewResolverFromCSV.
package macaddressouigo

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strings"
)

// Assignment describes one OUI entry.
type Assignment struct {
	// OUI is the 24-bit prefix, normalized to uppercase without separators,
	// e.g. "00000C".
	OUI string
	// Organization is the registered assignee name.
	Organization string
}

// Resolver looks up OUIs in an in-memory table.
type Resolver struct {
	table map[string]string
}

// Default returns a Resolver preloaded with a small, fixed subset of the
// IEEE OUI registry. The subset is intentionally tiny so the library stays
// self-contained; use NewResolverFromCSV for the full registry.
func Default() *Resolver {
	r := &Resolver{table: make(map[string]string)}
	// A handful of well-known, stable assignments. These are real IEEE
	// entries and are not expected to change.
	_ = r.Add("00000C", "Cisco Systems, Inc")
	_ = r.Add("00005E", "ICANN, IANA Department")
	_ = r.Add("000142", "Cisco Systems, Inc")
	_ = r.Add("0001C8", "Thomas Conrad Corp.")
	_ = r.Add("0002B3", "Intel Corporation")
	_ = r.Add("000347", "Intel Corporation")
	_ = r.Add("0004AC", "IBM Corp")
	_ = r.Add("000D3A", "Microsoft Corp.")
	_ = r.Add("001B63", "Apple, Inc.")
	_ = r.Add("001E52", "Apple, Inc.")
	_ = r.Add("002248", "Microsoft Corporation")
	_ = r.Add("0050C2", "IEEE Registration Authority")
	_ = r.Add("08002B", "DEC")
	_ = r.Add("FCFBFB", "Cisco Systems, Inc")
	return r
}

// NewResolver returns an empty Resolver.
func NewResolver() *Resolver {
	return &Resolver{table: make(map[string]string)}
}

// Add registers a single OUI assignment. The oui argument may contain
// separators (':', '-', '.') and any case; it is normalized before storage.
// It returns an error if the OUI is not a valid 24-bit prefix.
func (r *Resolver) Add(oui, organization string) error {
	key, err := normalizeOUI(oui)
	if err != nil {
		return err
	}
	if r.table == nil {
		r.table = make(map[string]string)
	}
	r.table[key] = organization
	return nil
}

// Lookup returns the organization registered for the OUI portion of mac.
// The second return value is false when the OUI is unknown or the input is
// not a well-formed MAC address.
func (r *Resolver) Lookup(mac string) (string, bool) {
	key, err := ouiFromMAC(mac)
	if err != nil {
		return "", false
	}
	org, ok := r.table[key]
	return org, ok
}

// Len reports the number of assignments held by the resolver.
func (r *Resolver) Len() int {
	return len(r.table)
}

// NewResolverFromCSV loads assignments from an IEEE-style CSV stream and
// returns a new Resolver. The reader must yield records whose first field is
// the OUI (with any of the usual separators) and whose second field is the
// organization name. Blank lines and lines starting with '#' are skipped.
//
// This exists so callers can supply the full public registry without the
// library itself depending on network access.
func NewResolverFromCSV(r io.Reader) (*Resolver, error) {
	res := NewResolver()
	scanner := bufio.NewScanner(r)
	// IEEE CSV files can contain long organization names; raise the buffer
	// limit above bufio's default 64 KiB line cap.
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		rec, err := csv.NewReader(strings.NewReader(line)).Read()
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNo, err)
		}
		if len(rec) < 2 {
			return nil, fmt.Errorf("line %d: expected at least 2 fields, got %d", lineNo, len(rec))
		}
		if err := res.Add(rec[0], rec[1]); err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNo, err)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

// normalizeOUI accepts an OUI in any of the common textual forms and returns
// the canonical six-hex-digit uppercase form.
func normalizeOUI(oui string) (string, error) {
	s := strings.ToUpper(strings.TrimSpace(oui))
	s = strings.NewReplacer(":", "", "-", "", ".", "").Replace(s)
	if len(s) != 6 {
		return "", fmt.Errorf("invalid OUI %q: expected 6 hex digits, got %d", oui, len(s))
	}
	for _, c := range s {
		if !isHex(c) {
			return "", fmt.Errorf("invalid OUI %q: non-hex character %q", oui, c)
		}
	}
	return s, nil
}

// ouiFromMAC extracts and normalizes the OUI from a MAC address string.
func ouiFromMAC(mac string) (string, error) {
	s := strings.TrimSpace(mac)
	if s == "" {
		return "", errors.New("empty MAC address")
	}
	s = strings.NewReplacer(":", "", "-", "", ".", "").Replace(s)
	if len(s) != 12 {
		return "", fmt.Errorf("invalid MAC %q: expected 12 hex digits, got %d", mac, len(s))
	}
	for _, c := range s {
		if !isHex(c) {
			return "", fmt.Errorf("invalid MAC %q: non-hex character %q", mac, c)
		}
	}
	return strings.ToUpper(s[:6]), nil
}

func isHex(c rune) bool {
	return (c >= '0' && c <= '9') || (c >= 'A' && c <= 'F') || (c >= 'a' && c <= 'f')
}
