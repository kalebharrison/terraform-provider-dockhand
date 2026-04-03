package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func testAccEnv(t *testing.T) (endpoint string, username string, password string) {
	t.Helper()

	endpoint = os.Getenv("DOCKHAND_TEST_ENDPOINT")
	username = os.Getenv("DOCKHAND_TEST_USERNAME")
	password = os.Getenv("DOCKHAND_TEST_PASSWORD")

	if endpoint == "" || username == "" || password == "" {
		t.Skip("acceptance test requires DOCKHAND_TEST_ENDPOINT, DOCKHAND_TEST_USERNAME, DOCKHAND_TEST_PASSWORD")
	}

	return endpoint, username, password
}

func testAccDefaultEnv() string {
	if v := os.Getenv("DOCKHAND_TEST_DEFAULT_ENV"); v != "" {
		return v
	}
	return "1"
}

func testAccConfigureProviderEnv(t *testing.T) string {
	t.Helper()

	endpoint, username, password := testAccEnv(t)
	t.Setenv("DOCKHAND_ENDPOINT", endpoint)
	t.Setenv("DOCKHAND_USERNAME", username)
	t.Setenv("DOCKHAND_PASSWORD", password)
	t.Setenv("DOCKHAND_DEFAULT_ENV", testAccDefaultEnv())

	if v := os.Getenv("DOCKHAND_TEST_AUTH_PROVIDER"); v != "" {
		t.Setenv("DOCKHAND_AUTH_PROVIDER", v)
	}

	return testAccDefaultEnv()
}

func testAccProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"dockhand": providerserver.NewProtocol6WithError(New("test")()),
	}
}

func testAccLoginSessionCookie(t *testing.T, endpoint string, username string, password string) string {
	t.Helper()

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("cookie jar: %v", err)
	}

	client := &http.Client{
		Jar: jar,
	}

	body := map[string]any{
		"username": username,
		"password": password,
		"provider": "local",
	}
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal login body: %v", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, endpoint+"/api/auth/login", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("login request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		t.Fatalf("login call: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode > 299 {
		t.Fatalf("login failed status=%d", res.StatusCode)
	}

	u, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		t.Fatalf("parse endpoint: %v", err)
	}

	cookies := jar.Cookies(u.URL)
	for _, c := range cookies {
		if c.Name == "dockhand_session" {
			// We pass Cookie header value directly through provider config.
			return fmt.Sprintf("%s=%s", c.Name, c.Value)
		}
	}

	t.Fatalf("dockhand_session cookie not found after login")
	return ""
}
