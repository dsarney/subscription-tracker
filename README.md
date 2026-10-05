# Subscription Tracker

A CLI for tracking subscriptions in a local Postgres database. You can list, add, update, and delete subscriptions (name, frequency, status, and auto-renew).

## Requirements

- [Go](https://go.dev/dl/)
- [Postgres](https://www.postgresql.org/download/)
- [goose](https://github.com/pressly/goose) (for migrations)

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

Ensure `$(go env GOPATH)/bin` is on your `PATH` so the `goose` command is available.

## Clone and set up

```bash
git clone https://github.com/dsarney/subscription-tracker
cd subscription-tracker
```

Create a Postgres database (name it `subscription_tracker` or match whatever you put in `.env`), then copy the example env file and set your connection string:

```bash
cp .env.example .env
```

Apply migrations:

```bash
export DATABASE_URL="$(grep DATABASE_URL .env | cut -d '=' -f2-)"
goose -dir migrations postgres "$DATABASE_URL" up
```

## Run

```bash
go run . list
go run . add Netflix monthly active true
go run . update <id> Netflix yearly inactive false
go run . delete <id>
```

| Command  | Arguments                                       |
| -------- | ----------------------------------------------- |
| `list`   | none                                            |
| `add`    | `<name> <frequency> <status> <auto_renew>`      |
| `update` | `<id> <name> <frequency> <status> <auto_renew>` |
| `delete` | `<id>`                                          |

- **frequency:** `monthly` or `yearly`
- **status:** `active` or `inactive`
- **auto_renew:** `true` or `false`

Names with spaces should be quoted, e.g. `go run . add "Disney Plus" monthly active true`.

## Run the tests

```bash
go test
```

Integration tests need Postgres and a valid `DATABASE_URL` (from `.env` or your environment). Tests that cannot connect are skipped.

## Build a binary (optional)

```bash
go build
```
