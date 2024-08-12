package dtaccess

import (
	"log"
	"time"
)

type ResourceData struct {
	Id                                                                           int32
	Resource_id, Resource_path, Manifest_file_name, Raw_file_path, Raw_file_name string
	Loaded_date                                                                  string
	Created_date                                                                 string
	Resource_status                                                              int8
}

func GetResourceById(resource_id string) (*ResourceData, error) {
	db, err := getDbConnection()
	if err != nil {
		log.Default().Println(err.Error())
		return nil, &DbConnectionError{}
	}

	defer db.Close()

	stmt, err := db.Prepare(`
		SELECT  
			id,
			resource_id,
			resource_path,
			manifest_file_name,
			raw_file_path,
			raw_file_name,
			COALESCE(loaded_date, ''),
			COALESCE(created_date, ''),
			resource_status
		FROM resources WHERE resource_id=?`)
	if err != nil {
		log.Default().Println(err.Error())
		return nil, &DbError{}
	}
	defer stmt.Close()

	rows, err := stmt.Query(resource_id)
	if err != nil {
		log.Default().Println(err.Error())
		return nil, &DbError{}
	}

	defer rows.Close()
	if rows.Next() {
		data := ResourceData{}
		err := rows.Scan(
			&data.Id,
			&data.Resource_id,
			&data.Resource_path,
			&data.Manifest_file_name,
			&data.Raw_file_path,
			&data.Raw_file_name,
			&data.Loaded_date,
			&data.Created_date,
			&data.Resource_status,
		)

		if err != nil {
			log.Default().Println(err.Error())
			return nil, &DbError{}
		}

		return &data, nil
	}

	return nil, &DbNotFound{}
}

func UpdateResource(rdata *ResourceData) error {
	db, err := getDbConnection()
	if err != nil {
		log.Default().Println(err.Error())
		return &DbConnectionError{}
	}

	defer db.Close()

	stmt, err := db.Prepare(`
		UPDATE resources 
		SET
			resource_path=?,
			manifest_file_name=?,
			raw_file_path=?,
			raw_file_name=?,
			loaded_date=?,
			resource_status=?
		WHERE resource_id=?
	`)
	if err != nil {
		log.Default().Println(err.Error())
		return &DbError{}
	}
	defer stmt.Close()

	result, err := stmt.Exec(
		rdata.Resource_path,
		rdata.Manifest_file_name,
		rdata.Raw_file_path,
		rdata.Raw_file_name,
		rdata.Loaded_date,
		rdata.Resource_status,
		rdata.Resource_id)

	if err != nil {
		log.Default().Println(err.Error())
		return &DbError{}
	}

	if rowsCount, err := result.RowsAffected(); rowsCount == 0 || err != nil {
		return &DbNotFound{}
	}

	return nil
}

func CreateResource(rdata *ResourceData) error {
	db, err := getDbConnection()
	if err != nil {
		log.Default().Println(err.Error())
		return &DbConnectionError{}
	}

	defer db.Close()

	stmt, err := db.Prepare(`
		INSERT INTO resources (
			resource_id,
			resource_path,
			manifest_file_name,
			raw_file_path,
			raw_file_name,
			created_date,
			resource_status)
		values (?,?,?,?,?,?,?)
	`)
	if err != nil {
		log.Default().Println(err.Error())
		return &DbError{}
	}
	defer stmt.Close()

	result, err := stmt.Exec(
		rdata.Resource_id,
		rdata.Resource_path,
		rdata.Manifest_file_name,
		rdata.Raw_file_path,
		rdata.Raw_file_name,
		time.Now().Format(time.RFC3339),
		rdata.Resource_status,
	)

	if err != nil {
		log.Default().Println(err.Error())
		return &DbError{}
	}

	id, err := result.LastInsertId()
	if err != nil {
		return &DbError{}
	}
	rdata.Id = int32(id)

	return nil
}

func GetResourcesToLoad() ([]ResourceData, error) {
	db, err := getDbConnection()
	if err != nil {
		log.Default().Println(err.Error())
		return nil, &DbConnectionError{}
	}

	defer db.Close()

	rows, err := db.Query(`
		SELECT  
			id,
			resource_id,
			resource_path,
			manifest_file_name,
			raw_file_path,
			raw_file_name,
			COALESCE(loaded_date, ''),
			COALESCE(created_date, ''),
			resource_status
		FROM resources 
		WHERE loaded_date is null 
		AND resource_status=2`)
	if err != nil {
		log.Default().Println(err.Error())
		return nil, &DbError{}
	}

	defer rows.Close()

	resourcesSlice := make([]ResourceData, 0)
	for rows.Next() {
		data := ResourceData{}
		err := rows.Scan(
			&data.Id,
			&data.Resource_id,
			&data.Resource_path,
			&data.Manifest_file_name,
			&data.Raw_file_path,
			&data.Raw_file_name,
			&data.Loaded_date,
			&data.Created_date,
			&data.Resource_status,
		)

		if err != nil {
			log.Default().Println(err.Error())
			return nil, &DbError{}
		}

		resourcesSlice = append(resourcesSlice, data)

	}

	return resourcesSlice, nil
}
