package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCommand(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantOutput string
		wantErr    bool
	}{
		{
			name:       "help flag shows usage",
			args:       []string{"--help"},
			wantOutput: "ankigo is a command-line interface",
			wantErr:    false,
		},
		{
			name:       "version command",
			args:       []string{"version"},
			wantOutput: "dev",
			wantErr:    false,
		},
		{
			name:       "version --full",
			args:       []string{"version", "--full"},
			wantOutput: "version: dev",
			wantErr:    false,
		},
		{
			name:       "deck create requires arg",
			args:       []string{"deck", "create"},
			wantOutput: "",
			wantErr:    true,
		},
		{
			name:       "note create requires flags",
			args:       []string{"note", "create"},
			wantOutput: "",
			wantErr:    true,
		},
		{
			name:       "note delete requires args",
			args:       []string{"note", "delete"},
			wantOutput: "",
			wantErr:    true,
		},
		{
			name:       "card search requires arg",
			args:       []string{"card", "search"},
			wantOutput: "",
			wantErr:    true,
		},
		{
			name:       "note list --help succeeds",
			args:       []string{"note", "list", "--help"},
			wantOutput: "List notes",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewRootCmd()
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			output := buf.String()
			if tt.wantOutput != "" && !strings.Contains(output, tt.wantOutput) {
				t.Errorf("Execute() output = %q, want to contain %q", output, tt.wantOutput)
			}
		})
	}
}
