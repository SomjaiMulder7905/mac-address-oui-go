# mac-address-oui-go

Resolves the Organizationally Unique Identifier (OUI) portion of a MAC address against an in-memory table of IEEE assignments.

```go
package main

import (
	"fmt"

	macaddressouigo "mac-address-oui-go"
)

func main() {
	r := macaddressouigo.Default()
	org, ok := r.Lookup("00:00:0C:01:02:03")
	if ok {
		fmt.Println(org) // Cisco Systems, Inc
	}
}
```

## Why this exists

The IEEE publishes the OUI registry as a large CSV that changes frequently. Shipping that file inside a library means the library goes stale, and downloading it at runtime means the library fails in offline or air-gapped environments. This package takes a third path: it embeds a small, hand-picked subset of stable assignments and exposes a CSV loader so callers who need the full registry can supply it themselves. The trade-off is coverage — `Default()` only knows about fourteen OUIs — in exchange for zero dependencies and no network access.

## API

- `Default() *Resolver` — returns a resolver preloaded with the embedded subset.
- `NewResolver() *Resolver` — returns an empty resolver.
- `(*Resolver).Add(oui, organization string) error` — registers one assignment. The OUI may be written with `:`, `-`, or `.` separators and in any case; it is normalized to six uppercase hex digits before storage. Re-adding an existing OUI overwrites it.
- `(*Resolver).Lookup(mac string) (string, bool)` — returns the organization for the OUI of `mac`. The boolean is false for unknown OUIs and for malformed inputs.
- `(*Resolver).Len() int` — number of assignments held.
- `NewResolverFromCSV(io.Reader) (*Resolver, error)` — loads an IEEE-style CSV stream. Blank lines and lines starting with `#` are skipped. Each record must have at least two fields: OUI and organization.

## The awkward edge

MAC addresses are conventionally written as six groups of two hex digits, but the IEEE registry and many tools also use the Cisco dotted form (`001b.63aa.bbcc`) or hyphenated form (`00-00-5E-AA-BB-CC`). `Lookup` accepts all three because rejecting them would surprise anyone reading a config file. What it does not accept is a MAC address with fewer or more than twelve hex digits, or one containing a non-hex character — those return `("", false)` rather than an error, since a miss is indistinguishable from an unknown OUI in most call sites. If you need to distinguish the two, validate the MAC before calling `Lookup`.
