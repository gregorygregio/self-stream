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

	defer db.Close()

	executeScripts(db)

	fmt.Println("InitDb finished..")
}

var scriptsMaps = make(map[string]string)

func loadScripts() {
	scriptsMaps["1_create_table_resources"] = `CREATE TABLE IF NOT EXISTS resources (
		id integer PRIMARY KEY,
		resource_id varchar(26) NOT NULL UNIQUE,
		resource_path varchar(4096) NOT NULL,
		manifest_file_name varchar(128) NOT NULL,
		raw_file_path varchar(4096) NOT NULL,
		raw_file_name varchar(128) NOT NULL,
		loaded_date INTEGER, 
		created_date INTEGER NOT NULL,
		resource_status TINYINT NOT NULL
	);`

	scriptsMaps["2_create_table_app_configurations"] = `CREATE TABLE IF NOT EXISTS app_configurations (
		config_name varchar(256) NOT NULL UNIQUE,
		data_type varchar(64) NOT NULL,
		config_value varchar(4096) NOT NULL,
		created_date INTEGER NOT NULL,
		updated_date INTEGER NOT NULL
	);`

	scriptsMaps["3_insert_appconfig_default_ingestfolder"] = `
		INSERT OR IGNORE INTO app_configurations (config_name, data_type, config_value, created_date, updated_date)
		VALUES ('ROOT_INGESTS_FOLDER', 'STRING', './resources/raw/', datetime(), datetime())
	`
	scriptsMaps["4_insert_appconfig_default_pkgs_folder"] = `
		INSERT OR IGNORE INTO app_configurations (config_name, data_type, config_value, created_date, updated_date)
		VALUES ('ROOT_PKGS_FOLDER', 'STRING', './resources/hls/', datetime(), datetime())
	`
}

func executeScripts(db *sql.DB) {

	loadScripts()
	fmt.Println("Executing scripts")

	for key, script := range scriptsMaps {

		fmt.Printf("Executing script: %v\n", key)
		_, err := db.Exec(script)
		onDbError(err)

		fmt.Printf("Script %v successfully executed\n", key)
	}

}

func onDbError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
