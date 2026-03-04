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
	if payload.ConnectionType != "hawser-edge" {
		t.Fatalf("expected agent connection type to map to hawser-edge, got %q", payload.ConnectionType)
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
		ConnectionType: "hawser-edge",
	}

	out := modelFromEnvironmentResponse(prior, resp)
	if out.ConnectionType.IsNull() || out.ConnectionType.ValueString() != "agent" {
		t.Fatalf("expected hawser-edge API type to normalize to agent state")
	}
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

func TestShouldProvisionAgentTokenCreateAndChangeOnly(t *testing.T) {
	createPlan := environmentModel{
		ConnectionType: types.StringValue("agent"),
		AgentToken:     types.StringValue("token-a"),
	}
	if !shouldProvisionAgentToken(createPlan, environmentModel{}) {
		t.Fatalf("expected create flow to provision agent token")
	}

	prior := environmentModel{
		ID:             types.StringValue("7"),
		ConnectionType: types.StringValue("agent"),
		AgentToken:     types.StringValue("token-a"),
	}
	samePlan := environmentModel{
		ConnectionType: types.StringValue("agent"),
		AgentToken:     types.StringValue("token-a"),
	}
	if shouldProvisionAgentToken(samePlan, prior) {
		t.Fatalf("expected unchanged token to skip reprovision")
	}

	changedPlan := environmentModel{
		ConnectionType: types.StringValue("agent"),
		AgentToken:     types.StringValue("token-b"),
	}
	if !shouldProvisionAgentToken(changedPlan, prior) {
		t.Fatalf("expected changed token to reprovision")
	}
}
