package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"

	// "hash"
	"time"

	"github.com/c-jamie/sql-manager-acc-auth/serverlib/internal/validator"
)

const (
	ScopeLogin = "login"
	ScopeRO    = "ro"
)
// Token defines the domain for the Token entity
type Token struct {
	Plaintext     string    `json:"plain_text"`
	Hash          []byte    `json:"hash"`
	UserAccountID int64     `json:"user_account_id"`
	Expiry        time.Time `json:"expiry"`
	Scope         string    `json:"scope"`
}

func generateToken(userID int64, ttl time.Duration, scope string) (*Token, error) {
	token := &Token{
		UserAccountID: userID,
		Expiry:        time.Now().Add(ttl),
		Scope:         scope,
	}

	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}
	if scope == ScopeLogin {
		token.Plaintext = "smt_" + base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)
	} else if scope == ScopeRO {
		token.Plaintext = "smr_" + base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)
	}
	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:]
	return token, nil
}

func ValidateTokenPlaintext(v *validator.Validator, tokenPlainText string) {
	v.Check(tokenPlainText != "", "token", "must be provided")
	v.Check(len(tokenPlainText) == 26+4, "token", "must be 30 bytes (chars) long")
}
// TokenModel wraps our connection pool
type TokenModel struct {
	DB *sql.DB
}

// New creates a new Token
func (m TokenModel) New(userID int64, ttl time.Duration, scope string) (*Token, error) {
	token, err := generateToken(userID, ttl, scope)
	if err != nil {
		return nil, err
	}

	err = m.Add(token)
	return token, err
}

// Add inserts a Token into the database
func (m TokenModel) Add(token *Token) error {
	query := `
		insert into token(hash, user_account_id, expiry, scope)
		values ($1, $2, $3, $4)
	`
	args := []interface{}{token.Hash, token.UserAccountID, token.Expiry, token.Scope}
	ctx, cancle := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancle()

	_, err := m.DB.ExecContext(ctx, query, args...)

	return err
}

// DeleteAllForUser removes all tokens for a UserAccount ID
func (m TokenModel) DeleteAllForUser(scope string, userID int64) error {
	query := `
		delete from token where scope = $1 and user_account_id = $1
	`
	args := []interface{}{scope, userID}
	ctx, cancle := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancle()

	_, err := m.DB.ExecContext(ctx, query, args...)

	return err
}
