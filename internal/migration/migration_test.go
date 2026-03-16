package migration

import "testing"

func TestParseNanosFromKey(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		want    int64
		wantErr bool
	}{
		{
			"valid key",
			"csp:1710629400123456789:a1b2c3d4e5f60718",
			1710629400123456789,
			false,
		},
		{
			"missing random part",
			"csp:1710629400123456789",
			0,
			true,
		},
		{
			"non-numeric nanos",
			"csp:notanumber:a1b2c3d4",
			0,
			true,
		},
		{
			"wrong prefix",
			"other:1710629400123456789:a1b2c3d4",
			0,
			true,
		},
		{
			"empty string",
			"",
			0,
			true,
		},
		{
			"only prefix",
			"csp:",
			0,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseNanosFromKey(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseNanosFromKey(%q) error = %v, wantErr %v", tt.key, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseNanosFromKey(%q) = %d, want %d", tt.key, got, tt.want)
			}
		})
	}
}
