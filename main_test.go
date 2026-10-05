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
