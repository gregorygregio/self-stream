package dtaccess

import "log"

type MovieData struct {
	Id                     int32
	Resource_id, File_name string
	Title                  string
	Year                   string
	Rated                  string
	Released               string
	Runtime                string
	Genre                  string
	Director               string
	Writer                 string
	Actors                 string
	Plot                   string
	Language               string
	Country                string
	Awards                 string
	Poster                 string
	Metascore              string
	Imdb_rating            string
	Imdb_votes             string
	Imdb_id                string
	Type                   string
	DVD                    string
	Production             string
	Website                string
	Loaded_Metadata        int8
}

func CreateMovie(movieData *MovieData) error {
	db, err := getDbConnection()
	if err != nil {
		log.Default().Println(err.Error())
		return &DbConnectionError{}
	}

	defer db.Close()

	stmt, err := db.Prepare(`
		INSERT INTO movies (
			resource_id,
			file_name,
			title,
			year,
			rated,
			released,
			runtime,
			genre,
			director,
			writer,
			actors,
			plot,
			language,
			country,
			awards,
			poster,
			metascore,
			imdb_rating,
			imdb_votes,
			imdb_id,
			type,
			dvd,
			production,
			website,
			loaded_Metadata
			)
		values (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
	`)
	if err != nil {
		log.Default().Println(err.Error())
		return &DbError{}
	}
	defer stmt.Close()

	result, err := stmt.Exec(
		movieData.Resource_id,
		movieData.File_name,
		movieData.Title,
		movieData.Year,
		movieData.Rated,
		movieData.Released,
		movieData.Runtime,
		movieData.Genre,
		movieData.Director,
		movieData.Writer,
		movieData.Actors,
		movieData.Plot,
		movieData.Language,
		movieData.Country,
		movieData.Awards,
		movieData.Poster,
		movieData.Metascore,
		movieData.Imdb_rating,
		movieData.Imdb_votes,
		movieData.Imdb_id,
		movieData.Type,
		movieData.DVD,
		movieData.Production,
		movieData.Website,
		movieData.Loaded_Metadata,
	)

	if err != nil {
		log.Default().Println(err.Error())
		return &DbError{}
	}

	id, err := result.LastInsertId()
	if err != nil {
		return &DbError{}
	}
	movieData.Id = int32(id)

	return nil
}

func GetMovieByFileName(fileName string) (*MovieData, error) {
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
			file_name,
			title,
			year,
			rated,
			released,
			runtime,
			genre,
			director,
			writer,
			actors,
			plot,
			language,
			country,
			awards,
			poster,
			metascore,
			imdb_rating,
			imdb_votes,
			imdb_id,
			type,
			dvd,
			production,
			website,
			loaded_Metadata
		FROM movies WHERE file_name = ?
	`)
	if err != nil {
		log.Default().Println(err.Error())
		return nil, &DbError{}
	}

	defer stmt.Close()

	rows, err := stmt.Query(fileName)
	if err != nil {
		log.Default().Println(err.Error())
		return nil, &DbError{}
	}

	defer rows.Close()

	if rows.Next() {
		data := MovieData{}
		err := rows.Scan(
			&data.Id,
			&data.Resource_id,
			&data.File_name,
			&data.Title,
			&data.Year,
			&data.Rated,
			&data.Released,
			&data.Runtime,
			&data.Genre,
			&data.Director,
			&data.Writer,
			&data.Actors,
			&data.Plot,
			&data.Language,
			&data.Country,
			&data.Awards,
			&data.Poster,
			&data.Metascore,
			&data.Imdb_rating,
			&data.Imdb_votes,
			&data.Imdb_id,
			&data.Type,
			&data.DVD,
			&data.Production,
			&data.Website,
			&data.Loaded_Metadata,
		)

		if err != nil {
			log.Default().Println(err.Error())
			return nil, &DbError{}
		}

		return &data, nil
	}

	return nil, &DbNotFound{}
}
