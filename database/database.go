package database

/*
import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"

	"hugpoint.tech/freebsd/forge/bugz"
	"hugpoint.tech/freebsd/forge/phab"
)

type DB struct {
	pool *sqlitex.Pool
}

func New(ctx context.Context, dbPath string) (*DB, error) {
	dbpool, err := sqlitex.NewPool(dbPath, sqlitex.PoolOptions{
		PoolSize: 10,
	})
	if err != nil {
		return nil, err
	}

	conn := dbpool.Get(ctx)
	if conn == nil {
		return nil, fmt.Errorf("new got nil conn")
	}
	defer dbpool.Put(conn)

	if err := createTables(conn); err != nil {
		return nil, err
	}

	return &DB{
		pool: dbpool,
	}, nil
}

func (db *DB) GetUsersTotal(ctx context.Context) (int, error) {
	conn := db.pool.Get(ctx)
	if conn == nil {
		return 0, fmt.Errorf("get total users got nil conn")
	}
	defer db.pool.Put(conn)
	defer fmt.Println()

	var totalUsersInDB int
	err := sqlitex.ExecuteTransient(conn, "SELECT count(id) from bugz_users;", &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			totalUsersInDB = stmt.ColumnInt(0)
			// fmt.Println("get total users totalBugsInDB", totalUsersInDB)
			return nil
		},
	})
	if err != nil {
		return 0, err
	}

	return totalUsersInDB, nil
}

func (db *DB) GetUsersMaxID(ctx context.Context) (int, error) {
	conn := db.pool.Get(ctx)
	if conn == nil {
		return 0, fmt.Errorf("get max user id - got nil conn")
	}
	defer db.pool.Put(conn)
	defer fmt.Println()

	var maxUserID int
	err := sqlitex.ExecuteTransient(conn, "SELECT max(id) from bugz_users;", &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			maxUserID = stmt.ColumnInt(0)
			return nil
		},
	})
	if err != nil {
		return 0, err
	}

	return maxUserID, nil
}

func (db *DB) GetUsers(ctx context.Context) ([]bugz.User, error) {
	conn := db.pool.Get(ctx)
	if conn == nil {
		return nil, fmt.Errorf("get bugs got nil conn")
	}
	defer db.pool.Put(conn)
	defer fmt.Println()

	result := []bugz.User{}
	err := sqlitex.ExecuteTransient(conn, `
	SELECT
		id,
		email,
		name,
		can_login,
		real_name
	FROM bugz_users;`, &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			// fmt.Println("get bugz_users id", stmt.ColumnInt(0))

			usr := bugz.User{
				ID:       stmt.ColumnInt(0),
				Email:    stmt.GetText("email"),
				Name:     stmt.GetText("name"),
				CanLogin: stmt.GetBool("can_login"),
				RealName: stmt.GetText("real_name"),
			}
			result = append(result, usr)

			return nil
		},
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (db *DB) GetUsersFromBugs(ctx context.Context, chanUsers chan bugz.User) error {
	conn := db.pool.Get(ctx)
	if conn == nil {
		return fmt.Errorf("get users from bugs - got nil conn")
	}
	defer db.pool.Put(conn)
	defer fmt.Println()

	filter := make(map[int]struct{})
	defer func() { fmt.Println("GetUsersFromBugs filter", len(filter)) }()

	err := sqlitex.ExecuteTransient(conn, `SELECT creator_detail FROM bugs;`, &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			cd := stmt.GetText("creator_detail")

			var bu bugz.UserDetail
			if err := json.Unmarshal([]byte(cd), &bu); err != nil {
				return err
			}

			usr := bugz.User{
				ID:       bu.ID,
				Email:    bu.Email,
				Name:     bu.Name,
				CanLogin: true,
				RealName: bu.RealName,
				// Groups:   []bugz.Group{},
			}

			if _, ok := filter[usr.ID]; !ok {
				// fmt.Println("GetUsersFromBugs creator_detail", cd)
				chanUsers <- usr
			}
			filter[usr.ID] = struct{}{}

			return nil
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) CreateUsers(ctx context.Context, chanUsers chan bugz.User) error {
	conn := db.pool.Get(ctx)
	if conn == nil {
		return fmt.Errorf("create users got nil conn")
	}
	defer db.pool.Put(conn)
	defer fmt.Println()

	var i int
	for user := range chanUsers {
		i++
		fmt.Printf("\rInserting user into sql: %d user.ID %d", i, user.ID)

		err := sqlitex.ExecuteTransient(
			conn,
			`INSERT OR REPLACE INTO bugz_users (
				id,
				email,
				name,
				can_login,
				real_name
			) VALUES (?, ?, ?, ?, ?);`,
			&sqlitex.ExecOptions{
				Args: []any{
					user.ID,
					user.Email,
					user.Name,
					user.CanLogin,
					user.RealName,
				},
			},
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *DB) GetBugsTotal(ctx context.Context) (int, error) {
	conn := db.pool.Get(ctx)
	if conn == nil {
		return 0, fmt.Errorf("get total bugs got nil conn")
	}
	defer db.pool.Put(conn)
	defer fmt.Println()

	var totalBugsInDB int
	err := sqlitex.ExecuteTransient(conn, "SELECT count(id) from bugs;", &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			totalBugsInDB = stmt.ColumnInt(0)
			// fmt.Println("get total bugs totalBugsInDB", totalBugsInDB)
			return nil
		},
	})
	if err != nil {
		return 0, err
	}

	return totalBugsInDB, nil
}

func (db *DB) GetBugs(ctx context.Context, bugsChan chan bugz.Bug) error {
	conn := db.pool.Get(ctx)
	if conn == nil {
		return fmt.Errorf("get bugs got nil conn")
	}
	defer db.pool.Put(conn)
	defer fmt.Println()

	err := sqlitex.ExecuteTransient(conn, `
	SELECT
		id,
		dupe_of,
		product,
		status,
		priority,
		severity,
		component,
		platform,
		is_cc_accessible,
		creation_time,
		resolution,
		is_creator_accessible,
		version,
		summary,
		url,
		is_confirmed,
		last_change_time,
		creator,
		creator_detail,
		assigned_to
	FROM bugs;`, &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			// fmt.Println("get bugs id dupe_of", stmt.ColumnInt(0), stmt.ColumnInt(1))

			cd := stmt.GetText("creator_detail")
			var creatorDetail bugz.UserDetail
			if err := json.Unmarshal([]byte(cd), &creatorDetail); err != nil {
				return err
			}

			bug := bugz.Bug{
				ID:                  stmt.ColumnInt(0),
				Product:             stmt.GetText("product"),
				Status:              stmt.GetText("status"),
				Priority:            stmt.GetText("priority"),
				Severity:            stmt.GetText("severity"),
				Component:           stmt.GetText("component"),
				Platform:            stmt.GetText("platform"),
				IsCCAccessible:      stmt.GetBool("is_cc_accessible"),
				Blocks:              []int{},
				DependsOn:           []int{},
				CreationTime:        stmt.GetText("creation_time"),
				Resolution:          stmt.GetText("resolution"),
				SeeAlso:             []string{},
				Keywords:            []string{},
				IsCreatorAccessible: stmt.GetBool("is_creator_accessible"),
				Version:             stmt.GetText("version"),
				Summary:             stmt.GetText("summary"),
				URL:                 stmt.GetText("url"),
				Groups:              []string{},
				Alias:               []string{},
				IsConfirmed:         stmt.GetBool("is_confirmed"),
				LastChangeTime:      stmt.GetText("last_change_time"),
				Flags:               []interface{}{},
				DupeOf:              stmt.ColumnInt(1),
				Creator:             stmt.GetText("creator"),
				CreatorDetail:       creatorDetail,
				AssignedTo:          stmt.GetText("assigned_to"),
				AssignedToDetail:    bugz.UserDetail{},
				CC:                  []string{},
				CCDetail:            []bugz.UserDetail{},
			}

			bugsChan <- bug
			return nil
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) CreateBugs(ctx context.Context, chanBug <-chan bugz.Bug) error {
	conn := db.pool.Get(ctx)
	if conn == nil {
		return fmt.Errorf("create bugs got nil conn")
	}
	defer db.pool.Put(conn)
	defer fmt.Println()

	var i int
	for bug := range chanBug {
		i++
		fmt.Printf("\rInserting bug into sql: %d", i)

		bcd, err := json.Marshal(bug.CreatorDetail)
		if err != nil {
			return err
		}

		err = sqlitex.ExecuteTransient(
			conn,
			`INSERT OR REPLACE INTO bugs (
				id,
				product,
				status,
				priority,
				severity,
				component,
				platform,
				is_cc_accessible,
				creation_time,
				resolution,
				is_creator_accessible,
				version,
				summary,
				url,
				is_confirmed,
				last_change_time,
				dupe_of,
				creator,
				creator_detail,
				assigned_to
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`,
			&sqlitex.ExecOptions{
				Args: []any{
					bug.ID,
					bug.Product,
					bug.Status,
					bug.Priority,
					bug.Severity,
					bug.Component,
					bug.Platform,
					bug.IsCCAccessible,
					bug.CreationTime,
					bug.Resolution,
					bug.IsCreatorAccessible,
					bug.Version,
					bug.Summary,
					bug.URL,
					bug.IsConfirmed,
					bug.LastChangeTime,
					bug.DupeOf,
					bug.Creator,
					string(bcd),
					bug.AssignedTo,
				},
			},
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *DB) GetPhabricatorData(ctx context.Context) error {
	fmt.Println("Getting data from Phabricator")
	phab := phab.NewPhabClient()
	users := phab.UserSearch().Get()

	conn := db.pool.Get(ctx)
	if conn == nil {
		return fmt.Errorf("get phabricator data - got nil conn")
	}
	defer db.pool.Put(conn)
	defer fmt.Println()

	for i, user := range users {
		fmt.Printf("\rInserting phab user into sql: %d", i)

		err := sqlitex.ExecuteTransient(
			conn,
			`  INSERT INTO phab_users (
				id,
				type,
				phid,
				username,
				real_name,
				date_created,
				date_modified,
				roles,
				policy_edit,
				policy_view
			)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`,
			&sqlitex.ExecOptions{
				Args: []any{
					user.ID,
					user.Type,
					user.Phid,
					user.Fields.Username,
					user.Fields.RealName,
					user.Fields.DateCreated,
					user.Fields.DateModified,
					strings.Join(user.Fields.Roles, ","),
					user.Fields.Policy.Edit,
					user.Fields.Policy.View,
				},
			},
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// it's important to have no trailing whitespaces at the end of the stmt, because error will happen.
func createTables(conn *sqlite.Conn) error {
	err := sqlitex.ExecuteTransient(conn, `
	CREATE TABLE IF NOT EXISTS bugz_users (
		id INTEGER PRIMARY KEY,
		email TEXT,
		name TEXT,
		can_login INTEGER,
		real_name TEXT
	);`, nil)
	if err != nil {
		return err
	}

	// err = sqlitex.ExecuteTransient(conn, `DROP TABLE bugs;`, nil)
	// if err != nil {
	// 	return err
	// }

	err = sqlitex.ExecuteTransient(conn, `
	CREATE TABLE IF NOT EXISTS bugs (
		id INTEGER PRIMARY KEY,
		product TEXT,
		status TEXT,
		priority TEXT,
		severity TEXT,
		component TEXT,
		platform TEXT,
		is_cc_accessible INTEGER,
		creation_time TEXT,
		resolution TEXT,
		is_open INTEGER,
		is_creator_accessible INTEGER,
		version TEXT,
		summary TEXT,
		url TEXT,
		is_confirmed INTEGER,
		last_change_time TEXT,
		dupe_of INTEGER,
		creator TEXT,
		creator_detail TEXT,
		assigned_to TEXT
	);`, nil)
	if err != nil {
		return err
	}

	err = sqlitex.ExecuteTransient(conn, `
	CREATE TABLE IF NOT EXISTS phab_users (
		id              INTEGER PRIMARY KEY,
		type            TEXT,
		phid            TEXT,
		username        TEXT,
		real_name       TEXT,
		date_created    INTEGER,
		date_modified   INTEGER,
		roles           TEXT,
		policy_edit     TEXT,
		policy_view     TEXT
	);`, nil)
	if err != nil {
		return err
	}

	return nil
}*/
