package data

import (
	"database/sql"
	"time"

	"github.com/c-jamie/sql-manager-acc-auth/serverlib/internal/permission"
)

type Models struct {
	UserAccount interface {
		// Add adds a UserAccount to the database
		Add(*UserAccount) error
		// Get returns a UserAccount from a given ID
		Get(userID int64) (*UserAccount, error)
		// Delete removes a UserAccount entity from the database
		Delete(*UserAccount) (error)
		// GetByEmail returns as UserAccount for a given email
		GetByEmail(email string) (*UserAccount, error)
		// GetForToken returns a UserAccount for a given tokenScope
		GetForToken(tokenScope string, token string) (*UserAccount, error)
		// Update updates a UserAccount entity 
		Update(*UserAccount) error
	}
	Token interface {
		// New creates a new Token
		New(userID int64, ttl time.Duration, scope string) (*Token, error)
		// Add inserts a Token into the database
		Add(token *Token) (error)
		// DeleteAllForUser removes all tokens for a UserAccount ID
		DeleteAllForUser(scope string, userID int64) error
	}
	Permission interface {
		GetForUser(userID int64) (Permissions, error)
	}
	Team interface {
		Add(team *Team) error
		Get(id int64) (*Team, error)
		Update(team *Team) error
	}
}

func NewModels(db *sql.DB) Models {
	manager := permission.PermissionManager{}
	return Models{
		UserAccountModel{DB: db},
		TokenModel{DB:db},
		PermissionModel{Manager: manager, DB: db},
		TeamModel{DB: db},
	}
}
