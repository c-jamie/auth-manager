package data

import (
	"database/sql"
	_ "embed"
	"fmt"

	pmanager "github.com/c-jamie/sql-manager-acc-auth/serverlib/internal/permission"
)
// Permissions contains all permissions for a given role
type Permissions []string
// PermissionModel wraps our connection pool
type PermissionModel struct {
	Manager pmanager.PermissionManager
	DB      *sql.DB
}
// Include checks for a permission code 
func (p Permissions) Include(code string) bool {
	for i := range p {
		if code == p[i] {
			return true
		}
	}
	return false

}
// GetForUser loads permissions for a given UserAccount ID
func (app PermissionModel) GetForUser(userID int64) (Permissions, error) {

	userMod := UserAccountModel{DB: app.DB}
	user, err := userMod.Get(userID)
	if err != nil {
		return nil, err
	}

	if user.Role == "" {
		return nil, fmt.Errorf("user has no role")
	}
	perm, err := app.Manager.GetForRole(user.Role)
	if err != nil {
		return nil, err
	}
	var permissions Permissions
	for _, p := range perm {
		permissions = append(permissions, p)
	}
	return permissions, nil
}
