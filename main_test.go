package main

import (
	"testing"
)

// TestConnectionDB checks that we can open and ping Postgres using DATABASE_URL.
// Skips when the DSN is unavailable (e.g. CI without a database).
func TestConnectionDB(t *testing.T) {
	dsn, err := loadDSN()
	if err != nil {
		t.Skipf("skipping test: %v", err)
	}

	db, err := connectionDB(dsn)
	if err != nil {
		t.Fatalf("connectionDB() error: %v", err)
	}
	defer db.Close()
}

// TestLoadDSN verifies a set DATABASE_URL is returned successfully.
func TestLoadDSN(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")

	dsn, err := loadDSN()
	if err != nil {
		t.Fatalf("loadDSN() error: %v", err)
	}
	if dsn == "" {
		t.Fatal("loadDSN() returned empty DSN")
	}
}

// TestLoadDSN_Missing verifies an empty DATABASE_URL produces an error.
func TestLoadDSN_Missing(t *testing.T) {
	t.Setenv("DATABASE_URL", "")

	_, err := loadDSN()
	if err == nil {
		t.Fatal("loadDSN() expected an error when DATABASE_URL is empty")
	}
}

// TestListSubscriptions ensures listing succeeds (empty table is valid).
func TestListSubscriptions(t *testing.T) {
	dsn, err := loadDSN()
	if err != nil {
		t.Skipf("skipping test: %v", err)
	}

	db, err := connectionDB(dsn)
	if err != nil {
		t.Fatalf("connectionDB() error: %v", err)
	}
	defer db.Close()

	subscriptions, err := listSubscriptions(db)
	if err != nil {
		t.Fatalf("listSubscriptions() error: %v", err)
	}

	if subscriptions == nil {
		t.Fatal("listSubscriptions() returned nil slice")
	}
}

// TestAddSubscription inserts a row, asserts it appears in list, then cleans up.
// defer order matters: cleanup delete must run before db.Close().
func TestAddSubscription(t *testing.T) {
	dsn, err := loadDSN()
	if err != nil {
		t.Skipf("skipping test: %v", err)
	}

	db, err := connectionDB(dsn)
	if err != nil {
		t.Fatalf("connectionDB() error: %v", err)
	}

	name := "Test Subscription"
	frequency := "monthly"
	status := "active"
	autoRenew := true

	defer db.Close()
	defer func() {
		if _, err := db.Exec(`DELETE FROM subscriptions WHERE name = $1`, name); err != nil {
			t.Errorf("cleanup delete: %v", err)
		}
	}()

	if err := addSubscription(db, name, frequency, status, autoRenew); err != nil {
		t.Fatalf("addSubscription() error: %v", err)
	}

	subscription, err := listSubscriptions(db)
	if err != nil {
		t.Fatalf("listSubscriptions() error: %v", err)
	}

	found := false
	for _, s := range subscription {
		if s.Name == name {
			found = true
			if s.Frequency != frequency {
				t.Fatalf("frequency  = %q, want %q", s.Frequency, frequency)
			}
			if s.Status != status {
				t.Fatalf("status  = %q, want %q", s.Status, status)
			}
			if s.AutoRenew != autoRenew {
				t.Fatalf("autoRenew  = %v, want %v", s.AutoRenew, autoRenew)
			}
			if s.ID == "" {
				t.Fatal("expected postgres to generate an ID")
			}
			break
		}
	}

	if !found {
		t.Fatalf("added subscription %q not found in list", name)
	}
}

// TestUpdateSubscription inserts a row, updates it by id, asserts new values, then cleans up by id.
func TestUpdateSubscription(t *testing.T) {
	dsn, err := loadDSN()
	if err != nil {
		t.Skipf("skipping test: %v", err)
	}

	db, err := connectionDB(dsn)
	if err != nil {
		t.Fatalf("connectionDB() error: %v", err)
	}

	originalName := "Test Update Subscription"
	updatedName := "Updated Subscription"
	updatedFrequency := "yearly"
	updatedStatus := "inactive"
	updatedAutoRenew := false

	var id string

	defer db.Close()
	defer func() {
		if id == "" {
			return
		}
		if _, err := db.Exec("DELETE FROM subscriptions WHERE id = $1", id); err != nil {
			t.Errorf("cleanup delete: %v", err)
		}
	}()

	if err := addSubscription(db, originalName, "monthly", "active", true); err != nil {
		t.Fatalf("addSubscription() error: %v", err)
	}

	subscriptions, err := listSubscriptions(db)
	if err != nil {
		t.Fatalf("listSubscriptions() error: %v", err)
	}

	for _, s := range subscriptions {
		if s.Name == originalName {
			id = s.ID
			break
		}
	}

	if id == "" {
		t.Fatalf("could not find added subscription %q", originalName)
	}

	if err := updateSubscription(db, id, updatedName, updatedFrequency, updatedStatus, updatedAutoRenew); err != nil {
		t.Fatalf("updateSubscription() error: %v", err)
	}

	subscriptions, err = listSubscriptions(db)
	if err != nil {
		t.Fatalf("listSubscriptions() error: %v", err)
	}

	found := false
	for _, s := range subscriptions {
		if s.ID == id {
			found = true
			if s.Name != updatedName {
				t.Fatalf("name = %q, want %q", s.Name, updatedName)
			}
			if s.Frequency != updatedFrequency {
				t.Fatalf("frequency = %q, want %q", s.Frequency, updatedFrequency)
			}
			if s.Status != updatedStatus {
				t.Fatalf("status = %q, want %q", s.Status, updatedStatus)
			}
			if s.AutoRenew != updatedAutoRenew {
				t.Fatalf("autoRenew = %v, want %v", s.AutoRenew, updatedAutoRenew)
			}
			break
		}
	}

	if !found {
		t.Fatalf("updated subscription %q not found in list", id)
	}
}

// TestDeleteSubscription inserts a row, deletes it by id, and asserts it no longer appears in list.
func TestDeleteSubscription(t *testing.T) {
	dsn, err := loadDSN()
	if err != nil {
		t.Skipf("skipping test: %v", err)
	}

	db, err := connectionDB(dsn)
	if err != nil {
		t.Fatalf("connectionDB() error: %v", err)
	}

	name := "Test Delete Subscription"
	var id string

	defer db.Close()
	defer func() {
		if id == "" {
			return
		}

		// Safety net if the test fails before deleteSubscription runs.
		if _, err := db.Exec("DELETE FROM subscriptions WHERE id = $1", id); err != nil {
			t.Errorf("cleanup delete: %v", err)
		}
	}()

	if err := addSubscription(db, name, "monthly", "active", true); err != nil {
		t.Fatalf("addSubscription() error: %v", err)
	}

	subscriptions, err := listSubscriptions(db)
	if err != nil {
		t.Fatalf("listSubscriptions() error: %v", err)
	}

	for _, s := range subscriptions {
		if s.Name == name {
			id = s.ID
			break
		}
	}
	if id == "" {
		t.Fatalf("could not find added subscription %q", name)
	}

	if err := deleteSubscription(db, id); err != nil {
		t.Fatalf("deleteSubscription() error: %v", err)
	}
	subscriptions, err = listSubscriptions(db)
	if err != nil {
		t.Fatalf("listSubscriptions() error: %v", err)
	}

	for _, s := range subscriptions {
		if s.ID == id {
			t.Fatalf("deleted subscription %q still found in list", id)
		}
	}
}
