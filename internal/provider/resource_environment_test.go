package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBuildEnvironmentPayloadIncludesAgentTokenForAgentConnection(t *testing.T) {
	plan := environmentModel{
		Name:           types.StringValue("agent-env"),
		ConnectionType: types.StringValue("agent"),
		AgentToken:     types.StringValue("agent-token-123"),
	}

	payload, err := buildEnvironmentPayload(plan, environmentModel{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.HawserToken == nil || *payload.HawserToken != "agent-token-123" {
		t.Fatalf("expected hawserToken in payload for agent connection")
	}
}

func TestBuildEnvironmentPayloadDoesNotIncludeAgentTokenForNonAgentConnection(t *testing.T) {
	plan := environmentModel{
		Name:           types.StringValue("direct-env"),
		ConnectionType: types.StringValue("direct"),
		AgentToken:     types.StringValue("should-not-be-sent"),
	}

	payload, err := buildEnvironmentPayload(plan, environmentModel{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.HawserToken != nil {
		t.Fatalf("expected hawserToken to be omitted for non-agent connection")
	}
}

func TestModelFromEnvironmentResponsePreservesPriorAgentTokenWhenRedacted(t *testing.T) {
	prior := environmentModel{
		AgentToken: types.StringValue("configured-token"),
	}
	resp := &environmentResponse{
		ID:             12,
		Name:           "agent-env",
		ConnectionType: "agent",
	}

	out := modelFromEnvironmentResponse(prior, resp)
	if out.AgentToken.IsNull() || out.AgentToken.ValueString() != "configured-token" {
		t.Fatalf("expected prior agent_token to be preserved when API omits token")
	}
}

func TestModelFromEnvironmentResponseClearsAgentTokenForNonAgentConnection(t *testing.T) {
	prior := environmentModel{
		AgentToken: types.StringValue("configured-token"),
	}
	resp := &environmentResponse{
		ID:             13,
		Name:           "direct-env",
		ConnectionType: "direct",
	}

	out := modelFromEnvironmentResponse(prior, resp)
	if !out.AgentToken.IsNull() {
		t.Fatalf("expected agent_token to be null for non-agent connection")
	}
}
