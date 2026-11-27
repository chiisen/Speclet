package version

import "testing"

func TestFormat(t *testing.T) {
	tests := []struct {
		name     string
		tag      string
		revision string
		want     string
	}{
		{name: "release tag", tag: "v0.1.0", revision: "1234567890abcdef", want: "v0.1.0"},
		{name: "development revision", tag: "dev", revision: "1234567890abcdef", want: "dev+1234567"},
		{name: "no revision", tag: "dev", want: "dev"},
		{name: "short revision", tag: "dev", revision: "123456", want: "dev"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := format(tt.tag, tt.revision); got != tt.want {
				t.Fatalf("format(%q, %q) = %q, want %q", tt.tag, tt.revision, got, tt.want)
			}
		})
	}
}
