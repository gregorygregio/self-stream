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
		config_name varchar(255) NOT NULL UNIQUE,
		data_type varchar(64) NOT NULL,
		config_value varchar(4096) NOT NULL,
		created_date INTEGER NOT NULL,
		updated_date INTEGER NOT NULL
	);`

	scriptsMaps["3_create_table_movies"] = `CREATE TABLE IF NOT EXISTS movies (
		id integer PRIMARY KEY,
		resource_id varchar(26) NOT NULL UNIQUE,
		file_name varchar(255),
		title VARCHAR(255),
		year VARCHAR(4),
		rated VARCHAR(5),
		released VARCHAR(15),
		runtime VARCHAR(15),
		genre VARCHAR(255),
		director VARCHAR(128),
		writer VARCHAR(255),
		actors VARCHAR(255),
		plot VARCHAR(512),
		language VARCHAR(128),
		country VARCHAR(128),
		awards VARCHAR(128),
		poster VARCHAR(255),
		metascore VARCHAR(10),
		imdb_rating VARCHAR(10),
		imdb_votes VARCHAR(20),
		imdb_id VARCHAR(20),
		type VARCHAR(64),
		dvd VARCHAR(64),
		box_office VARCHAR(32),
		production VARCHAR(128),
		website VARCHAR(255),
		loaded_metadata INTEGER NOT NULL
	);`

	scriptsMaps["4_insert_appconfig_default_pkgs_folder"] = `
		INSERT OR IGNORE INTO app_configurations (config_name, data_type, config_value, created_date, updated_date)
		VALUES ('ROOT_PKGS_FOLDER', 'STRING', './resources/hls/', datetime(), datetime())
	`

	// scriptsMaps["5_insert_appconfig_default_ingestfolder"] = `
	// 	INSERT OR IGNORE INTO app_configurations (config_name, data_type, config_value, created_date, updated_date)
	// 	VALUES ('ROOT_INGESTS_FOLDER', 'STRING', './resources/raw/', datetime(), datetime())
	// `
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
