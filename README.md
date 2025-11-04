# mini-ledger

A simple double-entry bookkeeping database with cryptographic verification built in Go

## Features

- **Double-entry accounting**: All transactions must balance (debits = credits)
- **REST API**: Create accounts, post transactions, view balances
- **Cryptographic verification**: Uses Merkle trees to ensure transaction integrity
- **SQLite database**: Lightweight persistent storage

## API Endpoints

- `POST /accounts` - Create new account
- `POST /transactions` - Post balanced transaction
- `GET /accounts/all` - View all account balances  
- `GET /transactions/all` - View transaction history
- `GET /verify` - Verify ledger integrity

## Quick Start

```bash
go run src/main.go
```

Server runs on `localhost:8080`

## Example Transaction

```json
{
  "description": "Payment to supplier",
  "entries": [
    {"account": "Cash", "amount": -1000},
    {"account": "Accounts Payable", "amount": 1000}
  ]
}
```

Built with Go 1.24, SQLite, and SHA-256 hashing for tamper detection.
