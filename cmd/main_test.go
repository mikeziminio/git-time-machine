package main

import (
	"bytes"
	"fmt"
	"os"
	"testing"
)

func TestRun_ArgumentValidation(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "unknown flag",
			args:     []string{"--unknown", "value"},
			wantErr:  true,
			errMsg:   "Unknown flag: --unknown",
		},
		{
			name:     "unknown flag short",
			args:     []string{"-x"},
			wantErr:  true,
			errMsg:   "Unknown flag: -x",
		},
		{
			name:     "missing input flag with argument",
			args:     []string{"myrepo"},
			wantErr:  true,
			errMsg:   "Unknown argument: myrepo (did you forget '-i' or '--input'?)",
		},
		{
			name:     "time-to without dash",
			args:     []string{"time-to", "9", "-i", "/input", "-o", "/output"},
			wantErr:  true,
			errMsg:   "Unknown argument: time-to (did you forget '-i' or '--input'?)",
		},
		{
			name:     "date-from without dash",
			args:     []string{"date-from", "2023-01-01", "-i", "/input", "-o", "/output"},
			wantErr:  true,
			errMsg:   "Unknown argument: date-from (did you forget '-i' or '--input'?)",
		},
		{
			name:     "valid flag after unknown",
			args:     []string{"--unknown", "--input", "/input", "-o", "/output"},
			wantErr:  true,
			errMsg:   "Unknown flag: --unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origArgs := os.Args
			defer func() { os.Args = origArgs }()

			os.Args = append([]string{"git-time-machine"}, tt.args...)

			var buf bytes.Buffer
			fmt.Fprintf(&buf, "")

			err := run()

			if (err != nil) != tt.wantErr {
				t.Errorf("run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil {
				if tt.errMsg != "" {
					if err.Error() != tt.errMsg {
						t.Errorf("run() error = %q, want %q", err.Error(), tt.errMsg)
					}
				}
			}
		})
	}
}
