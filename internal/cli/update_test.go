package cli

import "testing"

func TestNewerVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		candidate string
		current   string
		want      bool
	}{
		{name: "new patch", candidate: "v1.2.4", current: "v1.2.3", want: true},
		{name: "new minor", candidate: "v1.3.0", current: "v1.2.9", want: true},
		{name: "new major", candidate: "v2.0.0", current: "v1.9.9", want: true},
		{name: "same", candidate: "v1.2.3", current: "v1.2.3", want: false},
		{name: "older", candidate: "v1.2.2", current: "v1.2.3", want: false},
		{name: "invalid candidate", candidate: "latest", current: "v1.2.3", want: false},
		{name: "development build", candidate: "v1.2.3", current: "devel", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := newerVersion(tt.candidate, tt.current); got != tt.want {
				t.Fatalf("newerVersion(%q, %q) = %v, want %v", tt.candidate, tt.current, got, tt.want)
			}
		})
	}
}
