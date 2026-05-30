package sources

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var validTypes = map[string]bool{
	"channel": true, "supergroup": true, "group": true, "user": true, "bot": true,
}

// parseTypeFilter expands a slice of flag values (each possibly comma-separated)
// into a flat slice of validated type names. Returns an error if any value is
// not recognised.
func parseTypeFilter(flags []string) ([]string, error) {
	var out []string
	for _, f := range flags {
		for _, part := range strings.Split(f, ",") {
			s := strings.TrimSpace(part)
			if s == "" {
				continue
			}
			if !validTypes[s] {
				return nil, fmt.Errorf("unknown type %q (valid: channel, supergroup, group, user, bot)", s)
			}
			out = append(out, s)
		}
	}
	return out, nil
}

// outputFormat returns the value of the persistent --output flag (json|text).
func outputFormat(cmd *cobra.Command) string {
	if f := cmd.Flag("output"); f != nil {
		return f.Value.String()
	}
	return "json"
}
