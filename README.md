# gopg

A learning project written in Go to understand how a PostgreSQL client works.
It connects to PostgreSQL over TCP and builds and reads protocol messages directly.

## What it does

- Builds a startup message with a user and database name.
- Connects to a local PostgreSQL server.
- Reads startup responses and waits for the server to be ready.
- Sends simple SQL queries.
- Reads column information and row values, including NULL values.

## Requirements

- Go 1.27.1 or later.
- Docker with Docker Compose.

## Run the project

Start PostgreSQL from the project folder:

```sh
docker compose up -d
```

Once PostgreSQL is ready, run the Go program:

```sh
go run .
```

The program connects to `127.0.0.1:15432`. Both the user and database are named
`postgres`. The local Docker setup allows connections without a password.
Use this setup only for local development.

The SQL query is set in `main.go`. The current query, `select * from;`, is invalid
and is used to explore server responses. To try a valid query, change it to:

```go
query := "SELECT 1 AS number;"
```

The program prints the number of columns and rows in the result.

To stop PostgreSQL:

```sh
docker compose down
```

The database data stays in a Docker volume for the next run.

## Current limits

This project is still in progress.

- Only connections that need no password are supported.
- Connections use plain TCP without TLS.
- Some query decoding errors are not handled yet.
- Server errors are printed but are not returned as Go errors.
- The command tag is not saved in the result yet.
