package provider

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

type fakeGitStackDestroyClient struct {
	calls                []string
	deleteStackStatus    int
	deleteStackErr       error
	getStackFound        bool
	getStackErr          error
	deleteGitStackStatus int
	deleteGitStackErr    error
}

func (f *fakeGitStackDestroyClient) DeleteStack(_ context.Context, env string, name string) (int, error) {
	f.calls = append(f.calls, "DeleteStack:"+env+":"+name)
	return f.deleteStackStatus, f.deleteStackErr
}

func (f *fakeGitStackDestroyClient) GetStackByName(_ context.Context, env string, name string) (*stackResponse, bool, error) {
	f.calls = append(f.calls, "GetStackByName:"+env+":"+name)
	if f.getStackFound {
		return &stackResponse{Name: name}, true, f.getStackErr
	}
	return nil, false, f.getStackErr
}

func (f *fakeGitStackDestroyClient) DeleteGitStack(_ context.Context, env string, id string) (int, error) {
	f.calls = append(f.calls, "DeleteGitStack:"+env+":"+id)
	return f.deleteGitStackStatus, f.deleteGitStackErr
}

func TestBuildGitStackPayloadHonorsDeployTriggers(t *testing.T) {
	plan := gitStackModel{
		StackName:                 types.StringValue("test-stack"),
		ComposePath:               types.StringValue("stacks/app/compose.yml"),
		WebhookEnabled:            types.BoolValue(false),
		WebhookSecretAutoGenerate: types.BoolValue(false),
		WebhookSecret:             types.StringNull(),
		AutoUpdateEnabled:         types.BoolValue(false),
		AutoUpdateCron:            types.StringValue("0 3 * * *"),
		DeployNow:                 types.BoolValue(true),
		ForceRedeploy:             types.BoolValue(true),
		EnvVarsJSON:               types.StringValue("[]"),
		URL:                       types.StringValue("https://example.com/repo.git"),
		Branch:                    types.StringValue("main"),
	}

	payload, err := buildGitStackPayload(plan, gitStackDeployTriggers{DeployNow: true, ForceRedeploy: false})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !payload.DeployNow {
		t.Fatal("expected deployNow in payload")
	}
	if payload.ForceRedeploy {
		t.Fatal("expected forceRedeploy omitted when trigger false")
	}
}

func TestBuildGitStackPayloadIncludesContextDir(t *testing.T) {
	plan := gitStackModel{
		StackName:                 types.StringValue("test-stack"),
		ComposePath:               types.StringValue("stacks/metube/compose.yml"),
		ContextDir:                types.StringValue("."),
		EnvFilePath:               types.StringValue("./shared.env"),
		WebhookEnabled:            types.BoolValue(false),
		WebhookSecretAutoGenerate: types.BoolValue(false),
		WebhookSecret:             types.StringNull(),
		AutoUpdateEnabled:         types.BoolValue(false),
		AutoUpdateCron:            types.StringValue("0 3 * * *"),
		DeployNow:                 types.BoolValue(false),
		EnvVarsJSON:               types.StringValue("[]"),
		URL:                       types.StringValue("https://example.com/repo.git"),
		Branch:                    types.StringValue("main"),
	}

	payload, err := buildGitStackPayload(plan, gitStackDeployTriggers{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.ContextDir == nil || *payload.ContextDir != "." {
		t.Fatalf("expected contextDir . in payload, got %#v", payload.ContextDir)
	}
	if payload.EnvFilePath == nil || *payload.EnvFilePath != "./shared.env" {
		t.Fatalf("expected envFilePath ./shared.env in payload, got %#v", payload.EnvFilePath)
	}
}

func TestGitStackPayloadContextDirJSONTag(t *testing.T) {
	field, ok := reflect.TypeOf(gitStackPayload{}).FieldByName("ContextDir")
	if !ok {
		t.Fatal("expected ContextDir field on gitStackPayload")
	}
	if tag := field.Tag.Get("json"); tag != "contextDir,omitempty" {
		t.Fatalf("expected json tag contextDir,omitempty, got %q", tag)
	}
}

func TestModelFromGitStackResponseMapsContextDir(t *testing.T) {
	contextDir := "."
	composePath := "stacks/metube/compose.yml"
	resp := &gitStackResponse{
		ID:          1,
		StackName:   "test-stack",
		AutoUpdate:  false,
		ContextDir:  &contextDir,
		ComposePath: &composePath,
	}

	model := modelFromGitStackResponse(resp)
	if model.ContextDir.IsNull() || model.ContextDir.ValueString() != "." {
		t.Fatalf("expected context_dir . from response, got null=%v value=%q", model.ContextDir.IsNull(), model.ContextDir.ValueString())
	}
}

func TestMergeGitStackStatePreservesContextDirWhenRemoteOmitsIt(t *testing.T) {
	preferred := gitStackModel{
		ContextDir: types.StringValue("."),
	}
	remote := gitStackModel{
		ContextDir: types.StringNull(),
	}

	merged := mergeGitStackState(preferred, remote)
	if merged.ContextDir.IsNull() || merged.ContextDir.ValueString() != "." {
		t.Fatalf("expected context_dir to preserve configured value, got null=%v value=%q", merged.ContextDir.IsNull(), merged.ContextDir.ValueString())
	}
}

func TestBuildGitStackPayloadWebhookRequiresSecretOrAutoGenerate(t *testing.T) {
	plan := gitStackModel{
		StackName:                 types.StringValue("test-stack"),
		ComposePath:               types.StringValue("docker-compose.yml"),
		WebhookEnabled:            types.BoolValue(true),
		WebhookSecretAutoGenerate: types.BoolValue(false),
		WebhookSecret:             types.StringNull(),
		AutoUpdateEnabled:         types.BoolValue(false),
		AutoUpdateCron:            types.StringValue("0 3 * * *"),
		DeployNow:                 types.BoolValue(false),
		EnvVarsJSON:               types.StringValue("[]"),
		URL:                       types.StringValue("https://example.com/repo.git"),
		Branch:                    types.StringValue("main"),
	}

	_, err := buildGitStackPayload(plan, gitStackDeployTriggers{})
	if err == nil {
		t.Fatal("expected error when webhook is enabled without secret or auto-generate")
	}
	if !strings.Contains(err.Error(), "webhook_secret") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPrepareGitStackWebhookSecretAutoGenerates(t *testing.T) {
	plan := gitStackModel{
		WebhookEnabled:            types.BoolValue(true),
		WebhookSecretAutoGenerate: types.BoolValue(true),
		WebhookSecret:             types.StringNull(),
	}
	if err := prepareGitStackWebhookSecret(&plan, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan.WebhookSecret.IsNull() || strings.TrimSpace(plan.WebhookSecret.ValueString()) == "" {
		t.Fatal("expected auto-generate to populate webhook_secret on plan")
	}

	payload, err := buildGitStackPayload(gitStackModel{
		StackName:                 types.StringValue("test-stack"),
		ComposePath:               types.StringValue("docker-compose.yml"),
		WebhookEnabled:            types.BoolValue(true),
		WebhookSecretAutoGenerate: types.BoolValue(true),
		WebhookSecret:             plan.WebhookSecret,
		AutoUpdateEnabled:         types.BoolValue(false),
		AutoUpdateCron:            types.StringValue("0 3 * * *"),
		DeployNow:                 types.BoolValue(false),
		EnvVarsJSON:               types.StringValue("[]"),
		URL:                       types.StringValue("https://example.com/repo.git"),
		Branch:                    types.StringValue("main"),
	}, gitStackDeployTriggers{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.WebhookSecret == nil || *payload.WebhookSecret == "" {
		t.Fatal("expected generated webhook secret to be sent in payload")
	}
}

func TestPrepareGitStackWebhookSecretReusesState(t *testing.T) {
	plan := gitStackModel{
		WebhookEnabled:            types.BoolValue(true),
		WebhookSecretAutoGenerate: types.BoolValue(true),
		WebhookSecret:             types.StringNull(),
	}
	state := gitStackModel{WebhookSecret: types.StringValue("existing-secret")}
	if err := prepareGitStackWebhookSecret(&plan, &state); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan.WebhookSecret.ValueString() != "existing-secret" {
		t.Fatalf("expected state secret reuse, got %q", plan.WebhookSecret.ValueString())
	}
}

func TestBuildGitStackPayloadWebhookExplicitSecret(t *testing.T) {
	plan := gitStackModel{
		StackName:                 types.StringValue("test-stack"),
		ComposePath:               types.StringValue("docker-compose.yml"),
		WebhookEnabled:            types.BoolValue(true),
		WebhookSecretAutoGenerate: types.BoolValue(false),
		WebhookSecret:             types.StringValue("custom-secret"),
		AutoUpdateEnabled:         types.BoolValue(false),
		AutoUpdateCron:            types.StringValue("0 3 * * *"),
		DeployNow:                 types.BoolValue(false),
		EnvVarsJSON:               types.StringValue("[]"),
		URL:                       types.StringValue("https://example.com/repo.git"),
		Branch:                    types.StringValue("main"),
	}

	payload, err := buildGitStackPayload(plan, gitStackDeployTriggers{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.WebhookSecret == nil || *payload.WebhookSecret != "custom-secret" {
		t.Fatalf("expected explicit webhook secret to be sent")
	}
}

func TestBuildGitStackPayloadSetsBothAutoUpdateKeys(t *testing.T) {
	plan := gitStackModel{
		StackName:                 types.StringValue("test-stack"),
		ComposePath:               types.StringValue("docker-compose.yml"),
		WebhookEnabled:            types.BoolValue(false),
		WebhookSecretAutoGenerate: types.BoolValue(false),
		WebhookSecret:             types.StringNull(),
		AutoUpdateEnabled:         types.BoolValue(true),
		AutoUpdateCron:            types.StringValue("15 6 * * *"),
		DeployNow:                 types.BoolValue(false),
		EnvVarsJSON:               types.StringValue("[]"),
		URL:                       types.StringValue("https://example.com/repo.git"),
		Branch:                    types.StringValue("main"),
	}

	payload, err := buildGitStackPayload(plan, gitStackDeployTriggers{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !payload.AutoUpdateEnabled {
		t.Fatalf("expected AutoUpdateEnabled true in payload")
	}
	if payload.AutoUpdate == nil || !*payload.AutoUpdate {
		t.Fatalf("expected AutoUpdate pointer to be true in payload")
	}
}

func TestBuildGitStackPayloadIncludesDeployOptions(t *testing.T) {
	plan := gitStackModel{
		StackName:                 types.StringValue("test-stack"),
		ComposePath:               types.StringValue("docker-compose.yml"),
		WebhookEnabled:            types.BoolValue(false),
		WebhookSecretAutoGenerate: types.BoolValue(false),
		WebhookSecret:             types.StringNull(),
		AutoUpdateEnabled:         types.BoolValue(false),
		AutoUpdateCron:            types.StringValue("0 3 * * *"),
		DeployNow:                 types.BoolValue(true),
		BuildOnDeploy:             types.BoolValue(true),
		RepullImages:              types.BoolValue(false),
		ForceRedeploy:             types.BoolValue(true),
		EnvVarsJSON:               types.StringValue("[]"),
		URL:                       types.StringValue("https://example.com/repo.git"),
		Branch:                    types.StringValue("main"),
	}

	payload, err := buildGitStackPayload(plan, gitStackDeployTriggers{DeployNow: true, ForceRedeploy: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !payload.DeployNow {
		t.Fatalf("expected DeployNow true in payload")
	}
	if !payload.BuildOnDeploy {
		t.Fatalf("expected BuildOnDeploy true in payload")
	}
	if payload.RepullImages {
		t.Fatalf("expected RepullImages false in payload")
	}
	if !payload.ForceRedeploy {
		t.Fatalf("expected ForceRedeploy true in payload")
	}
}

func TestMergeGitStackStateWebhookSecretExplicitEmptyWins(t *testing.T) {
	preferred := gitStackModel{WebhookSecret: types.StringValue("")}
	remote := gitStackModel{WebhookSecret: types.StringValue("server-generated")}

	merged := mergeGitStackState(preferred, remote)
	if merged.WebhookSecret.IsNull() || merged.WebhookSecret.ValueString() != "" {
		t.Fatalf("expected explicit empty webhook_secret to win over server value")
	}
}

func TestMergeGitStackStateWebhookSecretUnsetInConfig(t *testing.T) {
	preferred := gitStackModel{
		WebhookSecret: types.StringNull(),
	}
	remote := gitStackModel{
		WebhookSecret: types.StringValue("server-generated-secret"),
	}

	merged := mergeGitStackState(preferred, remote)
	if !merged.WebhookSecret.IsNull() {
		t.Fatalf("expected webhook_secret to remain null when not configured, got %q", merged.WebhookSecret.ValueString())
	}
}

func TestMergeGitStackStateWebhookSecretConfiguredInConfig(t *testing.T) {
	preferred := gitStackModel{
		WebhookSecret: types.StringValue("from-config"),
	}
	remote := gitStackModel{
		WebhookSecret: types.StringNull(),
	}

	merged := mergeGitStackState(preferred, remote)
	if merged.WebhookSecret.IsNull() || merged.WebhookSecret.ValueString() != "from-config" {
		t.Fatalf("expected webhook_secret to preserve configured value, got null=%v value=%q", merged.WebhookSecret.IsNull(), merged.WebhookSecret.ValueString())
	}
}

func TestModelFromGitStackResponsePrefersAutoUpdateEnabled(t *testing.T) {
	falseVal := false
	resp := &gitStackResponse{
		ID:                1,
		StackName:         "test-stack",
		AutoUpdate:        true,
		AutoUpdateEnabled: &falseVal,
		WebhookEnabled:    false,
	}

	model := modelFromGitStackResponse(resp)
	if model.AutoUpdateEnabled.IsNull() || model.AutoUpdateEnabled.IsUnknown() {
		t.Fatalf("expected AutoUpdateEnabled to be set")
	}
	if model.AutoUpdateEnabled.ValueBool() {
		t.Fatalf("expected AutoUpdateEnabled=false when explicit autoUpdateEnabled is false")
	}
}

func TestModelFromGitStackResponseFallsBackToAutoUpdate(t *testing.T) {
	resp := &gitStackResponse{
		ID:                1,
		StackName:         "test-stack",
		AutoUpdate:        true,
		AutoUpdateEnabled: nil,
		WebhookEnabled:    false,
	}

	model := modelFromGitStackResponse(resp)
	if model.AutoUpdateEnabled.IsNull() || model.AutoUpdateEnabled.IsUnknown() {
		t.Fatalf("expected AutoUpdateEnabled to be set")
	}
	if !model.AutoUpdateEnabled.ValueBool() {
		t.Fatalf("expected AutoUpdateEnabled=true when falling back to autoUpdate")
	}
}

func TestModelFromGitStackResponseMapsDeployOptions(t *testing.T) {
	buildOnDeploy := true
	repullImages := false
	forceRedeploy := true
	resp := &gitStackResponse{
		ID:             1,
		StackName:      "test-stack",
		AutoUpdate:     true,
		WebhookEnabled: false,
		BuildOnDeploy:  &buildOnDeploy,
		RepullImages:   &repullImages,
		ForceRedeploy:  &forceRedeploy,
	}

	model := modelFromGitStackResponse(resp)
	if model.BuildOnDeploy.IsNull() || !model.BuildOnDeploy.ValueBool() {
		t.Fatalf("expected BuildOnDeploy=true from response")
	}
	if model.RepullImages.IsNull() || model.RepullImages.ValueBool() {
		t.Fatalf("expected RepullImages=false from response")
	}
	if model.ForceRedeploy.IsNull() || model.ForceRedeploy.ValueBool() {
		t.Fatalf("expected ForceRedeploy=false in state (one-shot flag not persisted from API)")
	}
}

func TestMergeGitStackStatePreservesDeployFlagsFromPreferred(t *testing.T) {
	preferred := gitStackModel{
		BuildOnDeploy: types.BoolValue(true),
		RepullImages:  types.BoolValue(false),
		DeployNow:     types.BoolValue(true),
		ForceRedeploy: types.BoolValue(true),
	}
	remote := gitStackModel{
		BuildOnDeploy: types.BoolNull(),
		RepullImages:  types.BoolNull(),
		ForceRedeploy: types.BoolValue(false),
		DeployNow:     types.BoolValue(false),
	}

	merged := mergeGitStackState(preferred, remote)
	if merged.BuildOnDeploy.IsNull() || !merged.BuildOnDeploy.ValueBool() {
		t.Fatalf("expected BuildOnDeploy to preserve configured true")
	}
	if merged.RepullImages.IsNull() || merged.RepullImages.ValueBool() {
		t.Fatalf("expected RepullImages to preserve configured false")
	}
	if !merged.DeployNow.ValueBool() {
		t.Fatalf("expected DeployNow=true from preferred config")
	}
	if !merged.ForceRedeploy.ValueBool() {
		t.Fatalf("expected ForceRedeploy=true from preferred config")
	}
}

func TestDestroyGitStackDeletesRuntimeBeforeGitRecord(t *testing.T) {
	client := &fakeGitStackDestroyClient{
		deleteStackStatus:    200,
		deleteGitStackStatus: 200,
	}

	err := destroyGitStack(context.Background(), client, "11", "77", "ollama")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantCalls := []string{
		"DeleteStack:11:ollama",
		"DeleteGitStack:11:77",
	}
	if !reflect.DeepEqual(client.calls, wantCalls) {
		t.Fatalf("unexpected call order: got=%v want=%v", client.calls, wantCalls)
	}
}

func TestDestroyGitStackSkipsGitDeleteWhenRuntimeStillExists(t *testing.T) {
	client := &fakeGitStackDestroyClient{
		deleteStackStatus: 500,
		deleteStackErr:    errors.New("runtime delete failed"),
		getStackFound:     true,
	}

	err := destroyGitStack(context.Background(), client, "11", "77", "ollama")
	if err == nil {
		t.Fatal("expected error")
	}

	wantCalls := []string{
		"DeleteStack:11:ollama",
		"GetStackByName:11:ollama",
	}
	if !reflect.DeepEqual(client.calls, wantCalls) {
		t.Fatalf("unexpected call order: got=%v want=%v", client.calls, wantCalls)
	}
}

func TestDestroyGitStackContinuesWhenRuntimeAlreadyGone(t *testing.T) {
	client := &fakeGitStackDestroyClient{
		deleteStackStatus:    500,
		deleteStackErr:       errors.New("runtime delete failed"),
		getStackFound:        false,
		deleteGitStackStatus: 200,
	}

	err := destroyGitStack(context.Background(), client, "11", "77", "ollama")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantCalls := []string{
		"DeleteStack:11:ollama",
		"GetStackByName:11:ollama",
		"DeleteGitStack:11:77",
	}
	if !reflect.DeepEqual(client.calls, wantCalls) {
		t.Fatalf("unexpected call order: got=%v want=%v", client.calls, wantCalls)
	}
}

func TestDestroyGitStackTreatsNotFoundAsSuccess(t *testing.T) {
	client := &fakeGitStackDestroyClient{
		deleteStackStatus:    404,
		deleteStackErr:       errors.New("not found"),
		deleteGitStackStatus: 404,
		deleteGitStackErr:    errors.New("not found"),
	}

	err := destroyGitStack(context.Background(), client, "11", "77", "ollama")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantCalls := []string{
		"DeleteStack:11:ollama",
		"DeleteGitStack:11:77",
	}
	if !reflect.DeepEqual(client.calls, wantCalls) {
		t.Fatalf("unexpected call order: got=%v want=%v", client.calls, wantCalls)
	}
}
