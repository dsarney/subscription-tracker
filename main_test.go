package main

import (
	"testing"
)

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

func TestLoadDSN_Missing(t *testing.T) {
	t.Setenv("DATABASE_URL", "")

	_, err := loadDSN()
	if err == nil {
		t.Fatal("loadDSN() expected an error when DATABASE_URL is empty")
	}
}

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
