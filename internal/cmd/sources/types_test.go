package sources

import (
	"reflect"
	"testing"
)

func TestParseTypeFilter(t *testing.T) {
	tests := []struct {
		in   []string
		want []string
		err  bool
	}{
		{nil, nil, false},
		{[]string{"channel"}, []string{"channel"}, false},
		{[]string{"channel,supergroup"}, []string{"channel", "supergroup"}, false},
		{[]string{"channel", "user,bot"}, []string{"channel", "user", "bot"}, false},
		{[]string{"weird"}, nil, true},
	}
	for _, tc := range tests {
		got, err := parseTypeFilter(tc.in)
		if (err != nil) != tc.err {
			t.Errorf("parseTypeFilter(%v) error = %v, wantErr = %v", tc.in, err, tc.err)
			continue
		}
		if !tc.err && !reflect.DeepEqual(got, tc.want) {
			t.Errorf("parseTypeFilter(%v) = %v, want %v", tc.in, got, tc.want)
		}
	}
}
