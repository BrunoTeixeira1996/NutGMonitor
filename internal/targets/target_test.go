package targets

import (
	"os"
	"testing"
)

func TestValidateSSHKeys(t *testing.T) {
	existingKey, err := os.CreateTemp(t.TempDir(), "key")
	if err != nil {
		t.Fatalf("failed to create temp key file: %v", err)
	}

	tests := []struct {
		name    string
		sshKey  string
		wantErr bool
	}{
		{name: "empty key path is always valid", sshKey: ""},
		{name: "existing key path is valid", sshKey: existingKey.Name()},
		{name: "missing key path is invalid", sshKey: "/nonexistent/path/does-not-exist", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tgt := Target{SSHKey: tt.sshKey}
			err := tgt.ValidateSSHKeys()
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSSHKeys() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
