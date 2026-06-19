package provider

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestGitCredentialPayloadUsesSSHPrivateKeyField(t *testing.T) {
	field, ok := reflect.TypeOf(gitCredentialPayload{}).FieldByName("SSHKey")
	if !ok {
		t.Fatal("expected SSHKey field on gitCredentialPayload")
	}
	if got := field.Tag.Get("json"); got != "sshPrivateKey,omitempty" {
		t.Fatalf("SSHKey json tag = %q, want sshPrivateKey,omitempty", got)
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
