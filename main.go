package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type Subscription struct {
	ID        string
	Name      string
	Frequency string
	Status    string
	AutoRenew bool
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: go run . <command>")
		fmt.Println("commands: list, add <name> <frequency> <status> <auto_renew>, update <id> <name> <frequency> <status> <auto_renew>, delete <id>")
		os.Exit(1)
	}

	command := strings.ToLower(os.Args[1])

	dsn, err := loadDSN()
	if err != nil {
		log.Fatalf("load DSN: %v", err)
	}

	db, err := connectionDB(dsn)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer db.Close()

	switch command {
	case "list":
		subscriptions, err := listSubscriptions(db)
		if err != nil {
			log.Fatalf("list subscriptions: %v", err)
		}

		if len(subscriptions) == 0 {
			fmt.Println("No subscriptions found")
			return
		}

		for _, s := range subscriptions {
			fmt.Printf("%s | %s | %s | %s | auto_renew=%v\n", s.ID, s.Name, s.Frequency, s.Status, s.AutoRenew)
		}
	case "add":
		if len(os.Args) < 6 {
			fmt.Println("usage: go run . add <name> <frequency> <status> <auto_renew>")
			os.Exit(1)
		}

		name := os.Args[2]
		frequency := strings.ToLower(os.Args[3])
		status := strings.ToLower(os.Args[4])

		autoRenew, err := strconv.ParseBool(os.Args[5])
		if err != nil {
			log.Fatalf("auto_renew must be true or false: %v", err)
		}

		if err := addSubscription(db, name, frequency, status, autoRenew); err != nil {
			log.Fatalf("add subscription: %v", err)
		}

		fmt.Println("subscription added")
	case "update":
		fmt.Println("update subscription")
	case "delete":
		fmt.Println("delete subscription")
	default:
		fmt.Printf("unknown command: %s\n", command)
		fmt.Println("commands: list, add <name> <frequency> <status> <auto_renew>, update <id> <name> <frequency> <status> <auto_renew>, delete <id>")
		os.Exit(1)
	}
}

func connectionDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	return db, nil
}

func loadDSN() (string, error) {
	_ = godotenv.Load()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return "", fmt.Errorf("DATABASE_URL is not set")
	}

	return dsn, nil
}

func listSubscriptions(db *sql.DB) ([]Subscription, error) {
	rows, err := db.Query(`SELECT id, name, frequency, status, auto_renew FROM subscriptions ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list subscriptions: %w", err)
	}
	defer rows.Close()

	subscriptions := []Subscription{}

	for rows.Next() {
		var s Subscription
		if err := rows.Scan(&s.ID, &s.Name, &s.Frequency, &s.Status, &s.AutoRenew); err != nil {
			return nil, fmt.Errorf("scan subscription: %w", err)
		}
		subscriptions = append(subscriptions, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return subscriptions, nil
}

func addSubscription(db *sql.DB, name string, frequency string, status string, autoRenew bool) error {
	_, err := db.Exec(`INSERT INTO subscriptions (name, frequency, status, auto_renew) 
	VALUES ($1, $2, $3, $4)`, name, frequency, status, autoRenew)

	if err != nil {
		return fmt.Errorf("insert subscription: %w", err)
	}

	return nil
}
