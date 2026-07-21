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

func TestBuildEnvironmentPayloadIncludesAgentTokenForAgentStandardConnection(t *testing.T) {
	plan := environmentModel{
		Name:           types.StringValue("agent-standard-env"),
		ConnectionType: types.StringValue("agent-standard"),
		AgentToken:     types.StringValue("agent-token-std"),
	}

	payload, err := buildEnvironmentPayload(plan, environmentModel{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.HawserToken == nil || *payload.HawserToken != "agent-token-std" {
		t.Fatalf("expected hawserToken in payload for agent-standard connection")
	}
	if payload.ConnectionType != "agent-standard" {
		t.Fatalf("expected agent-standard connection type to be preserved, got %q", payload.ConnectionType)
	}
}

func TestBuildEnvironmentPayloadDoesNotIncludeAgentTokenForNonAgentConnection(t *testing.T) {
	host := "docker.example.com"
	port := int64(2375)
	protocol := "http"
	plan := environmentModel{
		Name:           types.StringValue("direct-env"),
		ConnectionType: types.StringValue("direct"),
		Host:           types.StringValue(host),
		Port:           types.Int64Value(port),
		Protocol:       types.StringValue(protocol),
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

func TestModelFromEnvironmentResponsePreservesPriorAgentTokenForAgentStandard(t *testing.T) {
	prior := environmentModel{
		AgentToken: types.StringValue("configured-token"),
	}
	resp := &environmentResponse{
		ID:             14,
		Name:           "agent-std-env",
		ConnectionType: "agent-standard",
	}

	out := modelFromEnvironmentResponse(prior, resp)
	if out.ConnectionType.IsNull() || out.ConnectionType.ValueString() != "agent-standard" {
		t.Fatalf("expected agent-standard connection type to be preserved in state")
	}
	if out.AgentToken.IsNull() || out.AgentToken.ValueString() != "configured-token" {
		t.Fatalf("expected prior agent_token to be preserved for agent-standard connection")
	}
}

func TestModelFromEnvironmentResponsePreservesHawserStandardConnectionType(t *testing.T) {
	prior := environmentModel{
		ConnectionType: types.StringValue("hawser-standard"),
		AgentToken:     types.StringValue("configured-token"),
	}
	resp := &environmentResponse{
		ID:             15,
		Name:           "hawser-standard-env",
		ConnectionType: "hawser-standard",
	}

	out := modelFromEnvironmentResponse(prior, resp)
	if out.ConnectionType.IsNull() || out.ConnectionType.ValueString() != "hawser-standard" {
		t.Fatalf("expected hawser-standard connection type to be preserved in state")
	}
	if out.AgentToken.IsNull() || out.AgentToken.ValueString() != "configured-token" {
		t.Fatalf("expected prior agent_token to be preserved for hawser-standard")
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

func TestShouldProvisionAgentTokenForAgentStandard(t *testing.T) {
	createPlan := environmentModel{
		ConnectionType: types.StringValue("agent-standard"),
		AgentToken:     types.StringValue("token-a"),
	}
	if !shouldProvisionAgentToken(createPlan, environmentModel{}) {
		t.Fatalf("expected create flow to provision agent token for agent-standard")
	}

	prior := environmentModel{
		ID:             types.StringValue("7"),
		ConnectionType: types.StringValue("agent-standard"),
		AgentToken:     types.StringValue("token-a"),
	}
	samePlan := environmentModel{
		ConnectionType: types.StringValue("agent-standard"),
		AgentToken:     types.StringValue("token-a"),
	}
	if shouldProvisionAgentToken(samePlan, prior) {
		t.Fatalf("expected unchanged token to skip reprovision for agent-standard")
	}
}

func TestBuildEnvironmentPayloadIncludesPublicIP(t *testing.T) {
	ip := "203.0.113.10"
	host := "docker.example.com"
	plan := environmentModel{
		Name:           types.StringValue("edge"),
		ConnectionType: types.StringValue("direct"),
		Host:           types.StringValue(host),
		Port:           types.Int64Value(2375),
		Protocol:       types.StringValue("http"),
		PublicIP:       types.StringValue(ip),
	}

	payload, err := buildEnvironmentPayload(plan, environmentModel{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.PublicIP == nil || *payload.PublicIP != ip {
		t.Fatalf("expected publicIp %q in payload, got %#v", ip, payload.PublicIP)
	}
}

func TestModelFromEnvironmentResponsePublicIPDefaultsEmpty(t *testing.T) {
	resp := &environmentResponse{
		ID:             1,
		Name:           "socket-env",
		ConnectionType: "socket",
	}

	out := modelFromEnvironmentResponse(environmentModel{}, resp)
	if out.PublicIP.ValueString() != "" {
		t.Fatalf("expected empty public_ip default, got %q", out.PublicIP.ValueString())
	}
}

func TestModelFromEnvironmentResponsePublicIPFromAPI(t *testing.T) {
	ip := "198.51.100.4"
	resp := &environmentResponse{
		ID:             2,
		Name:           "remote-env",
		ConnectionType: "direct",
		PublicIP:       &ip,
	}

	out := modelFromEnvironmentResponse(environmentModel{}, resp)
	if out.PublicIP.ValueString() != ip {
		t.Fatalf("expected public_ip %q, got %q", ip, out.PublicIP.ValueString())
	}
}

func TestEnvironmentPublicIPNeedsFollowUp(t *testing.T) {
	ip := "10.1.7.185"
	empty := ""

	cases := []struct {
		name     string
		planIP   types.String
		response *environmentResponse
		want     bool
	}{
		{
			name:     "null plan skips",
			planIP:   types.StringNull(),
			response: &environmentResponse{},
			want:     false,
		},
		{
			name:     "unknown plan skips",
			planIP:   types.StringUnknown(),
			response: &environmentResponse{},
			want:     false,
		},
		{
			name:     "create omitted publicIp",
			planIP:   types.StringValue(ip),
			response: &environmentResponse{},
			want:     true,
		},
		{
			name:   "create returned empty publicIp",
			planIP: types.StringValue(ip),
			response: &environmentResponse{
				PublicIP: &empty,
			},
			want: true,
		},
		{
			name:   "create returned matching publicIp",
			planIP: types.StringValue(ip),
			response: &environmentResponse{
				PublicIP: &ip,
			},
			want: false,
		},
		{
			name:     "planned empty and create omitted matches",
			planIP:   types.StringValue(""),
			response: &environmentResponse{},
			want:     false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := environmentPublicIPNeedsFollowUp(environmentModel{PublicIP: tc.planIP}, tc.response)
			if got != tc.want {
				t.Fatalf("environmentPublicIPNeedsFollowUp() = %v, want %v", got, tc.want)
			}
		})
	}
}
