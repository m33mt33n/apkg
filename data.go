package apkg

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/adrg/xdg"
	_ "github.com/mattn/go-sqlite3"
)

type PackageList []string

var (
	Data_dir = filepath.Join(xdg.DataHome, "apkg")
	DB_path  = filepath.Join(Data_dir, "appdata.db")
)

func get_package_list(debug bool) (PackageList, uint8) {
	var list []string
	output, err := exec.Command("sh", "-c", "pm list packages").CombinedOutput()
	if debug {
		fmt.Println(string(output))
	}
	if err != nil {
		return list, 30
	}
	list = strings.Split(strings.Trim(string(output), "\n \t"), "\n")
	for idx, pkg := range list {
		list[idx] = pkg[8:]
	}
	return PackageList(list), 0
}

func print_count(msg string, count int) {
	if count == 0 {
		msg += "%d\n"
	} else {
		msg += "%02d\n"
	}
	fmt.Printf(msg, count)
}

func package_changed(db *sql.DB, pkg string) (bool, uint8) {
	var ver_name, ver_code string
	err := db.QueryRow("SELECT ver_name, ver_code FROM pkg_info WHERE pkg_name='"+pkg+"'").
		Scan(&ver_name, &ver_code)
	if err != nil {
		return false, 41
	}
	app, rcode := Get_info(Package(pkg))
	if rcode != 0 {
		return false, rcode
	}
	result := false
	if ver_name != app.Version || ver_code != app.Version_code {
		result = true
	}
	return result, 0
}

func Exec_add_record(stmt *sql.Stmt, pkg string) {
	app, _ := Get_info(Package(pkg))
	_, err := stmt.Exec(
		app.Package,
		app.Name,
		app.Version,
		app.Version_code,
		app.Main_activity,
		app.Minimum_SDK,
		app.Target_SDK,
		app.UID,
		app.First_installed,
		app.Last_update,
	)
	// better handling of this error
	if err != nil {
		fmt.Printf("> error: %s\n", err)
	}
}

func Exec_del_record(stmt *sql.Stmt, pkg string) {
	_, err := stmt.Exec(pkg)
	// better handling of this error
	if err != nil {
		fmt.Printf("> error: %s\n", err)
	}
}

func Create_db(new_db bool) uint8 {
	// create or remake database
	if new_db {
		fmt.Println("> making new database…")
	} else {
		fmt.Println("> remaking database…")
	}
	fmt.Println("> collecting data…")
	start := time.Now()
	package_list, rcode := get_package_list(false)
	if rcode != 0 {
		return rcode
	}
	os.MkdirAll(Data_dir, 0750)
	db, err := sql.Open("sqlite3", DB_path)
	if err != nil {
		return 31
	}
	defer db.Close()
	stmt0 := `
		CREATE TABLE IF NOT EXISTS pkg_info(
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			pkg_name TEXT,
			app_name TEXT,
			ver_name TEXT,
			ver_code TEXT,
			l_activity TEXT,
			min_sdk INTEGER,
			target_sdk INTEGER,
			uid INTEGER,
			first_install_time DATETIME,
			last_update_time DATETIME
    );
		DELETE FROM pkg_info;

		CREATE TABLE IF NOT EXISTS metadata(
			description TEXT PRIMARY KEY,
			timestamp DATETIME
    );
		DELETE FROM metadata;
	`
	_, err = db.Exec(stmt0)
	if err != nil {
		return 32
	}

	tx, err := db.Begin()
	if err != nil {
		return 33
	}
	stmt, err := tx.Prepare(`
		INSERT INTO pkg_info (
			pkg_name, app_name, ver_name, ver_code, l_activity, min_sdk, target_sdk, uid, first_install_time, last_update_time
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?
		);
	`)
	if err != nil {
		return 34
	}
	defer stmt.Close()
	pkg_count := len(package_list)
	for idx, pkg := range package_list {
		fmt.Printf("> processing… [package %02d of %d]\r", idx+1, pkg_count)
		Exec_add_record(stmt, pkg)
	}
	fmt.Println()
	err = tx.Commit()
	if err != nil {
		return 35
	}
	_, err = db.Exec(
		`INSERT INTO metadata VALUES("last_update", ?);`,
		time.Now().Round(time.Second),
	)
	if err != nil {
		return 36
	}
	t := time.Now()
	elapsed := t.Sub(start)
	fmt.Printf("> operation completed in %s\n", elapsed)
	return 0
}

func Check_db() uint8 {
	start := time.Now()
	fmt.Println("> checking database…")
	fmt.Println("> collecting data…")
	package_list, rcode := get_package_list(false)
	pkg_count_pm := len(package_list)
	if rcode != 0 {
		return rcode
	}
	db, err := sql.Open("sqlite3", DB_path)
	if err != nil {
		return 37
	}
	defer db.Close()
	var pkg_count_db int
	err = db.QueryRow("SELECT COUNT(pkg_name) FROM pkg_info;").Scan(&pkg_count_db)
	if err != nil {
		return 38
	}
	fmt.Printf("> installed packages count: %d\n", pkg_count_pm)
	fmt.Printf("> records found for packages: %d\n", pkg_count_db)
	fmt.Println("> started checking individual items…")
	var pkgs_in_db, pkgs_new, pkgs_removed, pkgs_minus_removed, pkgs_changed []string
	var pkgs_new_count, pkgs_removed_count, pkgs_changed_count int
	rows, err := db.Query("SELECT pkg_name FROM pkg_info;")
	if err != nil {
		return 39
	}
	defer rows.Close()
	var pkg_name string
	for rows.Next() {
		_ = rows.Scan(&pkg_name)
		pkgs_in_db = append(pkgs_in_db, pkg_name)
	}
	err = rows.Err()
	if err != nil {
		return 40
	}

	for _, pkg := range package_list {
		if !slices.Contains(pkgs_in_db, pkg) {
			pkgs_new = append(pkgs_new, pkg)
		}
	}
	pkgs_new_count = len(pkgs_new)
	print_count("> new packages found: ", pkgs_new_count)

	for _, pkg := range pkgs_in_db {
		if !slices.Contains(package_list, pkg) {
			pkgs_removed = append(pkgs_removed, pkg)
		} else {
			pkgs_minus_removed = append(pkgs_minus_removed, pkg)
		}
	}
	pkgs_removed_count = len(pkgs_removed)
	print_count("> packages uninstalled since last update: ", pkgs_removed_count)

	pkgs_to_check_count := len(pkgs_minus_removed)
	for idx, pkg := range pkgs_minus_removed {
		fmt.Printf("> checking for changed packages… [%02d/%d]\r", idx+1, pkgs_to_check_count)
		if changed, _ := package_changed(db, pkg); changed {
			pkgs_changed = append(pkgs_changed, pkg)
		}
	}
	fmt.Println()
	pkgs_changed_count = len(pkgs_changed)
	print_count("> packages changed since last update: ", pkgs_changed_count)

	if pkgs_new_count > 0 || pkgs_removed_count > 0 || pkgs_changed_count > 0 {
		if Confirm("would you like to update data?", false) {
			tx, err := db.Begin()
			if err != nil {
				return 42
			}
			stmt0, err := tx.Prepare(`
				INSERT INTO pkg_info (
					pkg_name, app_name, ver_name, ver_code, l_activity, min_sdk, target_sdk, uid, first_install_time, last_update_time
				) VALUES (
					?, ?, ?, ?, ?, ?, ?, ?, ?, ?
				);
			`)
			if err != nil {
				return 43
			}
			stmt1, err := tx.Prepare(
				`DELETE FROM pkg_info WHERE pkg_name=?`,
			)
			if err != nil {
				return 44
			}
			if pkgs_new_count > 0 {
				for _, pkg := range pkgs_new {
					Exec_add_record(stmt0, pkg)
				}
			}
			for _, pkg := range pkgs_removed {
				Exec_del_record(stmt1, pkg)
			}
			if pkgs_changed_count > 0 {
				for _, pkg := range pkgs_changed {
					Exec_del_record(stmt1, pkg)
					Exec_add_record(stmt0, pkg)
				}
			}
			err = tx.Commit()
			if err != nil {
				return 45
			}
			_, err = db.Exec(
				"UPDATE metadata SET timestamp=? WHERE description='last_update'",
				time.Now().Round(time.Second),
			)
			if err != nil {
				return 46
			}
		} else {
			return 130
		}
	} else {
		fmt.Println("> data is up-to-date, no action required!")
	}
	t := time.Now()
	elapsed := t.Sub(start)
	fmt.Printf("> operation completed in %s\n", elapsed)
	return 0
}

func Get_last_update_time() (time.Time, uint8) {
	var timestamp time.Time
	db, err := sql.Open("sqlite3", DB_path)
	if err != nil {
		return timestamp, 47
	}
	defer db.Close()
	err = db.QueryRow("SELECT timestamp FROM metadata WHERE description='last_update'").
		Scan(&timestamp)
	if err != nil {
		return timestamp, 48
	}
	return timestamp, 0
}

func Get_applist(use_pkgname bool, pattern string, pattern_is_regex bool, regex *regexp.Regexp) ([]App, uint8) {
	var applist []App
	db, err := sql.Open("sqlite3", DB_path)
	if err != nil {
		return applist, 49
	}
	defer db.Close()
	query := `
		SELECT
			pkg_name, app_name, ver_name, ver_code, l_activity, min_sdk, target_sdk, uid, first_install_time, last_update_time
		FROM
			pkg_info
	`
	criterion := "app_name"
	if use_pkgname {
		criterion = "pkg_name"
	}
	if !pattern_is_regex {
		query += fmt.Sprintf(" WHERE %s LIKE ?", criterion)
	}
	rows, err := db.Query(query, pattern)
	if err != nil {
		return applist, 50
	}
	defer rows.Close()
	var a App
	//a := new(App)
	for rows.Next() {
		_ = rows.Scan(
			&a.Package, &a.Name, &a.Version, &a.Version_code, &a.Main_activity,
			&a.Minimum_SDK, &a.Target_SDK, &a.UID, &a.First_installed, &a.Last_update,
		)
		if pattern_is_regex {
			criterion := a.Name
			if use_pkgname {
				criterion = string(a.Package)
			}
			if len(regex.FindString(criterion)) == 0 {
				continue
			}
		}
		applist = append(applist, a)
		//a = nil
	}
	err = rows.Err()
	if err != nil {
		return applist, 51
	}
	return applist, 0
}
