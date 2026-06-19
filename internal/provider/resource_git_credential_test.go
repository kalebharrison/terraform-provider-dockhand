package provider

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestGitCredentialPayloadUsesSSHPrivateKeyField(t *testing.T) {
	key := "-----BEGIN OPENSSH PRIVATE KEY-----\ntest\n-----END OPENSSH PRIVATE KEY-----"
	payload := gitCredentialPayload{
		Name:     "sample_key",
		AuthType: "ssh",
		SSHKey:   &key,
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	body := string(raw)
	if !strings.Contains(body, `"sshPrivateKey"`) {
		t.Fatalf("expected sshPrivateKey JSON field, got %s", body)
	}
	if strings.Contains(body, `"sshKey"`) {
		t.Fatalf("did not expect legacy sshKey JSON field, got %s", body)
	}
}

func TestBuildGitCredentialPayloadRequiresSSHKey(t *testing.T) {
	plan := gitCredentialModel{
		Name:     types.StringValue("sample_key"),
		AuthType: types.StringValue("ssh"),
	}

	_, err := buildGitCredentialPayload(plan, gitCredentialModel{}, true)
	if err == nil {
		t.Fatal("expected error when ssh_key missing on create")
	}
}
