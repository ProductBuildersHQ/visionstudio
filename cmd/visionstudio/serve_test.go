package main

import "testing"

func TestParseListenAddr(t *testing.T) {
	tests := []struct {
		name        string
		address     string
		port        int
		wantAddr    string
		wantNonLoop bool
		wantErr     bool
	}{
		{name: "default loopback from port", address: "", port: 9400, wantAddr: "127.0.0.1:9400"},
		{name: "explicit localhost", address: "localhost:9401", port: 9400, wantAddr: "localhost:9401"},
		{name: "explicit loopback ip", address: "127.0.0.1:9401", port: 9400, wantAddr: "127.0.0.1:9401"},
		{name: "port-only defaults host to loopback", address: ":9402", port: 9400, wantAddr: "127.0.0.1:9402"},
		{name: "non-loopback warns", address: "0.0.0.0:9403", port: 9400, wantAddr: "0.0.0.0:9403", wantNonLoop: true},
		{name: "non-loopback lan ip warns", address: "192.168.1.10:9404", port: 9400, wantAddr: "192.168.1.10:9404", wantNonLoop: true},
		{name: "missing port errors", address: "localhost", port: 9400, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, nonLoop, err := parseListenAddr(tt.address, tt.port)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseListenAddr(%q,%d) expected error, got addr %q", tt.address, tt.port, addr)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseListenAddr(%q,%d) unexpected error: %v", tt.address, tt.port, err)
			}
			if addr != tt.wantAddr {
				t.Errorf("addr = %q, want %q", addr, tt.wantAddr)
			}
			if nonLoop != tt.wantNonLoop {
				t.Errorf("nonLoopback = %v, want %v", nonLoop, tt.wantNonLoop)
			}
		})
	}
}

func TestIsLoopbackHost(t *testing.T) {
	loopback := []string{"localhost", "LOCALHOST", "127.0.0.1", "127.0.0.5", "::1"}
	for _, h := range loopback {
		if !isLoopbackHost(h) {
			t.Errorf("isLoopbackHost(%q) = false, want true", h)
		}
	}
	remote := []string{"0.0.0.0", "192.168.1.10", "example.com", "10.0.0.1"}
	for _, h := range remote {
		if isLoopbackHost(h) {
			t.Errorf("isLoopbackHost(%q) = true, want false", h)
		}
	}
}
