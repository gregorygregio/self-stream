package dtaccess

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func getDbConnection() (*sql.DB, error) {
	return sql.Open("sqlite3", "./resources/database/selfstream.db")
}

func InitDb() {
	fmt.Println("Starting InitDb")

	db, err := getDbConnection()
	onDbError(err)

	fmt.Println("Database connection successful")

	createTables(db)

	fmt.Println("InitDb finished..")
}

func createTables(db *sql.DB) {
	fmt.Println("creating resources table")

	sql := `CREATE TABLE IF NOT EXISTS resources (
		id integer PRIMARY KEY,
		resource_id varchar(256) NOT NULL UNIQUE,
		resource_path varchar(4096) NOT NULL,
		manifest_file_name varchar(128) NOT NULL,
		raw_file_path varchar(4096) NOT NULL,
		raw_file_name varchar(128) NOT NULL,
		loaded_date INTEGER, 
		created_date INTEGER NOT NULL
	);`

	result, err := db.Exec(sql)
	onDbError(err)

	rowsAffected, err := result.RowsAffected()
	onDbError(err)

	fmt.Printf("Resource table created: %v\n", rowsAffected)

}

func onDbError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
