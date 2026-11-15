package blockchain

import "testing"

func TestIsHexAddress(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0", true},
		{"0x742d35cc6634c0532925a3b844bc9e7595f0beb0", true},
		{"742d35Cc6634C0532925a3b844Bc9e7595f0bEb0", false},
		{"0x1234", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := IsHexAddress(tt.in); got != tt.want {
			t.Fatalf("IsHexAddress(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}
