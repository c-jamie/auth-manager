package data

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/c-jamie/sql-manager-acc-auth/serverlib/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

var AnonUser = &UserAccount{}
// UserAccount defines the domain for the user entity
type UserAccount struct {
	ID          int64             `json:"id"`
	Name        string            `json:"name"`
	Email       string            `json:"email"`
	Password    password          `json:"-"`
	Activated   bool              `json:"activated"`
	CreatedAt   time.Time         `json:"created_at"`
	Version     int               `json:"-"`
	Team        *Team             `json:"team"`
	Role        string            `json:"role"`
	Permissions map[string]string `json:"permissions"`
}
type password struct {
	plaintext *string
	hash      []byte
}
//UserAccountModel wraps our connection pool
type UserAccountModel struct {
	DB *sql.DB
}
// IsAnon returns true if the UserAccount is anonymous
func (m *UserAccount) IsAnon() bool {
	return (m == AnonUser)
}
// Add adds a UserAccount to the database
func (m UserAccountModel) Add(user *UserAccount) error {
	query := `
		with ins as (
			insert into team(name, created_at)
			values ($1, now())
			on conflict (name)
			do nothing
			returning id, created_at
		)
		select * from ins
		union
		select id, created_at
		from team where name = $1
	`

	args := []interface{}{user.Team.Name}
	ctx, cancle := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancle()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.Team.ID, &user.Team.CreatedAt)
	if err != nil {
		return err
	}

	query = `
		insert into user_account(email, password_hash, activated, role, created_at)
		values ($1, $2, $3, $4, now())
		returning id, created_at, version
	`
	args = []interface{}{user.Email, user.Password.hash, user.Activated, user.Role}
	ctx, cancle = context.WithTimeout(context.Background(), 3*time.Second)

	defer cancle()

	err = m.DB.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.CreatedAt, &user.Version)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "duplicate key value violates unique constraint"):
			return fmt.Errorf("duplicate email: %w", err)
		default:
			return err
		}
	}
	query = `
		insert into users_teams(user_account_id, team_id)
		values ($1, $2)
	`

	args = []interface{}{user.ID, user.Team.ID}
	ctx, cancle = context.WithTimeout(context.Background(), 3*time.Second)

	defer cancle()

	_, err = m.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}
// GetByEmail returns as UserAccount for a given email
func (m UserAccountModel) GetByEmail(email string) (*UserAccount, error) {
	query := `
		select 	ua.id
				, ua.created_at
				, ua.email
				, ua.password_hash
				, ua.activated
				, ua.version
				, ua.role
				, t.id
				, t.name
				, t.created_at
		from 	user_account as ua
		inner 	join users_teams as ut
		on		ua.id = ut.user_account_id
		inner 	join team as t
		on 		t.id = ut.team_id
		where 	ua.email = $1
	`

	var user UserAccount
	var team Team

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Version,
		&user.Role,
		&team.ID,
		&team.Name,
		&team.CreatedAt,
	)

	user.Team = &team

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, fmt.Errorf("no record found: %w", err)
		default:
			return nil, err
		}
	}
	return &user, nil
}

// Get returns a UserAccount from a given ID
func (m UserAccountModel) Get(userID int64) (*UserAccount, error) {
	query := `
		select 	ua.id
				, ua.created_at
				, ua.email
				, ua.password_hash
				, ua.activated
				, ua.version
				, ua.role
				, t.id
				, t.name
				, t.created_at
		from 	user_account as ua
		inner 	join users_teams as ut
		on		ua.id = ut.user_account_id
		inner 	join team as t
		on 		t.id = ut.team_id
		where 	ua.id = $1
	`

	var user UserAccount
	var team Team

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Version,
		&user.Role,
		&team.ID,
		&team.Name,
		&team.CreatedAt,
	)

	user.Team = &team

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, fmt.Errorf("no record found: %w", err)
		default:
			return nil, err
		}
	}
	return &user, nil
}

// Update updates a UserAccount entity 
func (m UserAccountModel) Update(user *UserAccount) error {
	query := `
		update 		user_account 
		set 		email = $1, password_hash = $2, activated = $3, role = $4, version = version + 1
		where 		id = $5 and version = $6
		returning 	version
	`
	args := []interface{}{
		user.Email,
		user.Password.hash,
		user.Activated,
		user.Role,
		user.ID,
		user.Version,
	}

	ctx, cancle := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancle()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.Version)

	if err != nil {
		switch {
		case strings.Contains(err.Error(), "duplicate key value violates unique constraint"):
			return fmt.Errorf("duplicate email: %w", err)
		case errors.Is(err, sql.ErrNoRows):
			return fmt.Errorf("conflict: %w", err)
		default:
			return err
		}
	}

	return nil
}

// Delete removes a UserAccount entity from the database
func (m UserAccountModel) Delete(user *UserAccount) error {
	query := `
		delete 		from user_account 
		where 		id = $1 and version = $2
		returning 	version
	`
	args := []interface{}{
		user.ID,
		user.Version,
	}

	ctx, cancle := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancle()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.Version)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return fmt.Errorf("unable to delete user: %w", err)
		default:
			return err
		}
	}

	return nil
}

// GetForToken returns a UserAccount for a given tokenScope
func (m UserAccountModel) GetForToken(tokenScope, tokenPlainText string) (*UserAccount, error) {
	tokenHash := sha256.Sum256([]byte(tokenPlainText))

	query := `
		select
				u.id
				, u.created_at
				, u.email
				, u.password_hash
				, u.activated
				, u.version
				, u.role
		from 	user_account as u
		inner join token as t
		on u.id = t.user_account_id
		where t.hash = $1
		and t.scope = $2
		and t.expiry > $3
	`
	args := []interface{}{tokenHash[:], tokenScope, time.Now()}

	var user UserAccount
	ctx, cancle := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancle()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Version,
		&user.Role,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, fmt.Errorf("no records %w", err)
		default:
			return nil, err
		}
	}
	return &user, nil
}
// Set adds a password to the password struct
func (p *password) Set(ptpassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(ptpassword), 13)
	if err != nil {
		return err
	}

	p.plaintext = &ptpassword
	p.hash = hash
	return nil
}
// Matches compares a plaintext password with the database
func (p *password) Matches(ptpassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(ptpassword))

	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			{
				return false, nil
			}
		default:
			return true, nil
		}
	}
	return true, nil
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
}

func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 chars long")
	v.Check(len(password) <= 72, "password", "must be less than 72 bytes (chars) long")
}

func ValidateUser(v *validator.Validator, user *UserAccount) {
	v.Check(user.Team.Name != "", "team", "must not be empty")
	ValidateEmail(v, user.Email)
	if user.Password.plaintext != nil {
		ValidatePasswordPlaintext(v, *user.Password.plaintext)
	}

	if user.Password.hash == nil {
		panic("missing password hash for user")
	}
}
