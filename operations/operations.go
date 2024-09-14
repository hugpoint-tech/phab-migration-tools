package operations

import (
	"hugpoint.tech/freebsd/forge/bugzilla"
	"hugpoint.tech/freebsd/forge/database"
	"hugpoint.tech/freebsd/forge/util"
)

/// This package contains high-level logic of user-facing operations.
/// This package is necessary to decouple other packages from each other
/// and keep main clean.

func DownloadBugzillaBugs(bugz *bugzilla.Client, db *database.DB) {
	bugs, err := bugz.DownloadBugzillaBugs()
	util.CheckFatal("failed to download bugzilla bugs", err)

	for _, bug := range bugs {
		err = db.InsertBug(bug)
		util.CheckFatal("failed to insert bug into the database", err)

	}
}
