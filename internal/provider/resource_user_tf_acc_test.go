package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccUserResourceTerraform(t *testing.T) {
	endpoint, username, password := testAccEnv(t)

	// Configure provider via env vars so configs can stay minimal.
	t.Setenv("DOCKHAND_ENDPOINT", endpoint)
	t.Setenv("DOCKHAND_USERNAME", username)
	t.Setenv("DOCKHAND_PASSWORD", password)
	t.Setenv("DOCKHAND_DEFAULT_ENV", "1")

	userName := "tf_acc_user_tf_" + time.Now().UTC().Format("20060102150405")
	pw := "TerraformAccTest123!"

	resourceName := "dockhand_user.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"dockhand": providerserver.NewProtocol6WithError(New("test")()),
		},
		CheckDestroy: testAccCheckUserDestroyed(endpoint, username, password),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig(userName, pw, "tf-acc-1@example.local", "TF Acc 1", false, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "username", userName),
					resource.TestCheckResourceAttr(resourceName, "email", "tf-acc-1@example.local"),
					resource.TestCheckResourceAttr(resourceName, "display_name", "TF Acc 1"),
					resource.TestCheckResourceAttr(resourceName, "is_admin", "false"),
					resource.TestCheckResourceAttr(resourceName, "is_active", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
			{
				Config: testAccUserConfig(userName, pw, "tf-acc-2@example.local", "TF Acc 2", true, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "email", "tf-acc-2@example.local"),
					resource.TestCheckResourceAttr(resourceName, "display_name", "TF Acc 2"),
					resource.TestCheckResourceAttr(resourceName, "is_admin", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password", "updated_at", "last_login"},
			},
		},
	})
}

func testAccUserConfig(username string, password string, email string, displayName string, isAdmin bool, isActive bool) string {
	// Provider uses DOCKHAND_* env vars.
	return fmt.Sprintf(`
provider "dockhand" {}

resource "dockhand_user" "test" {
  username     = %q
  password     = %q
  email        = %q
  display_name = %q
  is_admin     = %t
  is_active    = %t
}
`, username, password, email, displayName, isAdmin, isActive)
}

func testAccCheckUserDestroyed(endpoint string, username string, password string) func(state *terraform.State) error {
	return func(state *terraform.State) error {
		sessionCookie, err := testAccLoginSessionCookieForDestroy(endpoint, username, password)
		if err != nil {
			return err
		}
		client, err := NewClient(endpoint, sessionCookie, "1", true)
		if err != nil {
			return err
		}

		for _, rs := range state.RootModule().Resources {
			if rs.Type != "dockhand_user" {
				continue
			}

			_, status, err := client.GetUser(context.Background(), rs.Primary.ID)
			if err == nil {
				return fmt.Errorf("user still exists: id=%s", rs.Primary.ID)
			}
			if status != 404 {
				return fmt.Errorf("unexpected status checking user destroy: id=%s status=%d err=%v", rs.Primary.ID, status, err)
			}
		}

		return nil
	}
}

func testAccLoginSessionCookieForDestroy(endpoint string, username string, password string) (string, error) {
	// This is used inside terraform-plugin-testing callbacks, so avoid *testing.T helpers here.
	jar, err := cookiejar.New(nil)
	if err != nil {
		return "", fmt.Errorf("cookie jar: %w", err)
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
		return "", fmt.Errorf("marshal login body: %w", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, endpoint+"/api/auth/login", bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("login call: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode > 299 {
		return "", fmt.Errorf("login failed status=%d", res.StatusCode)
	}

	u, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("parse endpoint: %w", err)
	}

	cookies := jar.Cookies(u.URL)
	for _, c := range cookies {
		if c.Name == "dockhand_session" {
			// We pass Cookie header value directly through provider config.
			return fmt.Sprintf("%s=%s", c.Name, c.Value), nil
		}
	}

	return "", fmt.Errorf("dockhand_session cookie not found after login")
}
