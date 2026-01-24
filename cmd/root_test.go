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
			name:       "card add requires flags",
			args:       []string{"card", "add"},
			wantOutput: "",
			wantErr:    true,
		},
		{
			name:       "card search requires arg",
			args:       []string{"card", "search"},
			wantOutput: "",
			wantErr:    true,
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
