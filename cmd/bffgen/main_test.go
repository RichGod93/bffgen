package main

import (
	"runtime/debug"
	"testing"
)

func TestResolveVersion(t *testing.T) {
	info := func(v string) func() (*debug.BuildInfo, bool) {
		return func() (*debug.BuildInfo, bool) {
			return &debug.BuildInfo{Main: debug.Module{Version: v}}, true
		}
	}
	noInfo := func() (*debug.BuildInfo, bool) { return nil, false }

	tests := []struct {
		name    string
		ldflags string
		info    func() (*debug.BuildInfo, bool)
		want    string
	}{
		{"ldflags win", "v3.5.1", info("v3.5.0"), "v3.5.1"},
		{"go install", "dev", info("v3.5.1"), "v3.5.1"},
		{"source build", "dev", info("(devel)"), "dev"},
		{"no build info", "dev", noInfo, "dev"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveVersion(tt.ldflags, tt.info); got != tt.want {
				t.Errorf("resolveVersion() = %q, want %q", got, tt.want)
			}
		})
	}
}
