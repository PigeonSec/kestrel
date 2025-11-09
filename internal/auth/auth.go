package auth

import (
	"database/sql"
	"sync"
)

// Account represents a user account with API key access
type Account struct {
	APIKey string `json:"api_key"`
	Email  string `json:"email"`
	Plan   string `json:"plan"`
	Active bool   `json:"active"`
}

// KeyStore manages API keys in memory with SQLite persistence
type KeyStore struct {
	mu   sync.RWMutex
	keys map[string]*Account
	db   *sql.DB
}

// NewKeyStore creates a new key store with SQLite backend
func NewKeyStore(dbPath string) (*KeyStore, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Create table if not exists
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS accounts (
			api_key TEXT PRIMARY KEY,
			email TEXT,
			plan TEXT,
			active INTEGER DEFAULT 1
		);
	`)
	if err != nil {
		db.Close()
		return nil, err
	}

	ks := &KeyStore{
		keys: make(map[string]*Account),
		db:   db,
	}

	// Load existing keys from database
	if err := ks.loadFromDB(); err != nil {
		db.Close()
		return nil, err
	}

	return ks, nil
}

// IsValid checks if an API key is valid and active
func (ks *KeyStore) IsValid(apiKey string) bool {
	ks.mu.RLock()
	defer ks.mu.RUnlock()
	account, ok := ks.keys[apiKey]
	return ok && account.Active
}

// GetAccount retrieves an account by API key
func (ks *KeyStore) GetAccount(apiKey string) (*Account, bool) {
	ks.mu.RLock()
	defer ks.mu.RUnlock()
	account, ok := ks.keys[apiKey]
	return account, ok
}

// AddAccount adds or updates an account
func (ks *KeyStore) AddAccount(account *Account) error {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	// Update in-memory store
	ks.keys[account.APIKey] = account

	// Persist to database
	active := 0
	if account.Active {
		active = 1
	}
	_, err := ks.db.Exec(`
		INSERT OR REPLACE INTO accounts (api_key, email, plan, active)
		VALUES (?, ?, ?, ?)
	`, account.APIKey, account.Email, account.Plan, active)

	return err
}

// RemoveAccount removes an account
func (ks *KeyStore) RemoveAccount(apiKey string) error {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	delete(ks.keys, apiKey)
	_, err := ks.db.Exec(`DELETE FROM accounts WHERE api_key = ?`, apiKey)
	return err
}

// ListAccounts returns all accounts
func (ks *KeyStore) ListAccounts() []*Account {
	ks.mu.RLock()
	defer ks.mu.RUnlock()

	accounts := make([]*Account, 0, len(ks.keys))
	for _, account := range ks.keys {
		accounts = append(accounts, account)
	}
	return accounts
}

// ReplaceAll replaces all keys (used for external provider sync)
func (ks *KeyStore) ReplaceAll(accounts []*Account) error {
	newKeys := make(map[string]*Account, len(accounts))
	for _, acc := range accounts {
		if acc.Active && acc.APIKey != "" {
			newKeys[acc.APIKey] = acc
		}
	}

	ks.mu.Lock()
	ks.keys = newKeys
	ks.mu.Unlock()

	// Persist to database
	return ks.saveAllToDB(accounts)
}

func (ks *KeyStore) loadFromDB() error {
	rows, err := ks.db.Query(`SELECT api_key, email, plan, active FROM accounts`)
	if err != nil {
		return err
	}
	defer rows.Close()

	ks.mu.Lock()
	defer ks.mu.Unlock()

	for rows.Next() {
		var account Account
		var active int
		if err := rows.Scan(&account.APIKey, &account.Email, &account.Plan, &active); err != nil {
			continue
		}
		account.Active = active == 1
		if account.Active {
			ks.keys[account.APIKey] = &account
		}
	}
	return nil
}

func (ks *KeyStore) saveAllToDB(accounts []*Account) error {
	tx, err := ks.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Clear existing
	if _, err := tx.Exec(`DELETE FROM accounts`); err != nil {
		return err
	}

	// Insert all
	stmt, err := tx.Prepare(`INSERT INTO accounts (api_key, email, plan, active) VALUES (?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, acc := range accounts {
		if acc.APIKey == "" {
			continue
		}
		active := 0
		if acc.Active {
			active = 1
		}
		if _, err := stmt.Exec(acc.APIKey, acc.Email, acc.Plan, active); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// Close closes the database connection
func (ks *KeyStore) Close() error {
	return ks.db.Close()
}
