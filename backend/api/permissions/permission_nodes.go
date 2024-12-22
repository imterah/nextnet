package permissions

import "git.terah.dev/imterah/hermes/api/dbcore"

var DefaultPermissionNodes []string = []string{
	"routes.add",
	"routes.remove",
	"routes.start",
	"routes.stop",
	"routes.edit",
	"routes.visible",
	"routes.visibleConn",

	"backends.add",
	"backends.remove",
	"backends.start",
	"backends.stop",
	"backends.edit",
	"backends.visible",
	"backends.secretVis",

	"permissions.see",

	"users.add",
	"users.remove",
	"users.lookup",
	"users.edit",
}

func UserHasPermission(user *dbcore.User, node string) bool {
	for _, permission := range user.Permissions {
		if permission.PermissionNode == node && permission.HasPermission {
			return true
		}
	}

	return false
}
