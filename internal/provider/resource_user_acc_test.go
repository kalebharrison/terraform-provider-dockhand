package provider

import (
	"testing"
	"time"
)

func TestAccUserResource(t *testing.T) {
	endpoint, username, password := testAccEnv(t)
	sessionCookie := testAccLoginSessionCookie(t, endpoint, username, password)

	client, err := NewClient(endpoint, sessionCookie, "1", true)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	// Create
	userName := "tf_acc_user_" + time.Now().UTC().Format("20060102150405")
	pw := "TerraformAccTest123!"
	created, err := client.CreateUser(t.Context(), userPayload{
		Username: userName,
		Password: &pw,
		IsAdmin:  false,
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	// Read
	read, status, err := client.GetUser(t.Context(), intToString(created.ID))
	if err != nil {
		t.Fatalf("get user: status=%d err=%v", status, err)
	}
	if read.Username != userName {
		t.Fatalf("username mismatch: got=%q want=%q", read.Username, userName)
	}

	// Update
	email := "tf-acc@example.local"
	displayName := "TF Acc"
	updated, err := client.UpdateUser(t.Context(), intToString(created.ID), userPayload{
		Username:    userName,
		Email:       &email,
		DisplayName: &displayName,
		IsAdmin:     true,
		IsActive:    true,
	})
	if err != nil {
		t.Fatalf("update user: %v", err)
	}
	if updated.Email == nil || *updated.Email != email {
		t.Fatalf("email mismatch: got=%v want=%q", updated.Email, email)
	}
	if !updated.IsAdmin {
		t.Fatalf("expected isAdmin=true after update")
	}

	// Delete
	status, err = client.DeleteUser(t.Context(), intToString(created.ID))
	if err != nil && status != 404 {
		t.Fatalf("delete user: status=%d err=%v", status, err)
	}
	_, status, err = client.GetUser(t.Context(), intToString(created.ID))
	if err == nil || status != 404 {
		t.Fatalf("expected 404 after delete, got status=%d err=%v", status, err)
	}
}

func intToString(v int64) string {
	// Avoid strconv import in the test file; keep helpers local/simple.
	if v == 0 {
		return "0"
	}

	neg := false
	if v < 0 {
		neg = true
		v = -v
	}

	var buf [32]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + (v % 10))
		v /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
