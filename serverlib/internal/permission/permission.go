package permission

import (
	"embed"
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	casbin_fs_adapter "github.com/naucon/casbin-fs-adapter"
)

const (
	Read  = "read"
	Write = "write"
	SQLM  = "sqlm"
)

//go:embed *.txt
var casbinModel embed.FS
//go:embed *.csv
var casbinPolicy embed.FS

type PermissionManager struct {
}

func (pm *PermissionManager) GetForRole(role string) ([]string, error) {

	mf, err := casbinModel.ReadFile("casbin.txt")
	if err != nil {
		return nil, err
	}

	m, err := model.NewModelFromString(string(mf))
	if err != nil {
		return nil, err
	}

	policies := casbin_fs_adapter.NewAdapter(casbinPolicy, "policy.csv")
	enforcer, err := casbin.NewEnforcer(m, policies)
	if err != nil {
		return nil, err
	}
	perm := enforcer.GetFilteredPolicy(0, role)

	var permissions []string
	for _, p  := range perm {
		permissions = append(permissions, p[1] + "-" + p[2])
	}

	if err != nil {
		return nil, err
	}
	return permissions, nil
}
