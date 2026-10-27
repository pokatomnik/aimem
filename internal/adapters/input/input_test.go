package input

import (
	"strings"
	"testing"
)

func TestConfirm(t *testing.T) {
	tests := []struct {
		name     string
		response string
		want     bool
		wantErr  bool
	}{
		{name: "default yes", response: "\n", want: true},
		{name: "yes", response: "Y\n", want: true},
		{name: "yes word", response: " yes \n", want: true},
		{name: "no", response: "n\n", want: false},
		{name: "no word", response: "NO\n", want: false},
		{name: "invalid response", response: "maybe\n", wantErr: true},
		{name: "response before EOF", response: "y", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(strings.NewReader(tt.response)).Confirm("Continue?")
			if (err != nil) != tt.wantErr {
				t.Fatalf("Confirm() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("Confirm() = %v, want %v", got, tt.want)
			}
		})
	}
}
