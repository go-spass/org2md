package main

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestRootCmdVersion(t *testing.T) {
	cmd := newRootCmd(buildInfo{
		version: "1.2.3",
		commit:  "abc1234",
		date:    "2026-10-28T00:00:00Z",
	})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() = %v, want nil", err)
	}

	got := out.String()
	for _, want := range []string{"org2md", "1.2.3", "abc1234", "2026-10-28T00:00:00Z"} {
		if !strings.Contains(got, want) {
			t.Errorf("--version output is missing %q, got %q", want, got)
		}
	}
}

func TestRootCmdArgCount(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{name: "no arguments", args: nil, wantErr: true},
		{name: "two input files", args: []string{"a.org", "b.org"}, wantErr: true},
		{name: "one input file", args: []string{"a.org"}, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newRootCmd(buildInfo{version: "test"})
			cmd.SetOut(io.Discard)
			cmd.SetErr(io.Discard)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if tt.wantErr && err == nil {
				t.Fatal("Execute() = nil, want an argument error")
			}
			if !tt.wantErr && err != nil && !strings.Contains(err.Error(), "not implemented") {
				t.Fatalf("Execute() = %v, want nil or a not-implemented error", err)
			}
		})
	}
}
