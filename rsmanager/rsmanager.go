package rsmanager

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"self-stream/appconfigs"
	"self-stream/dtaccess"
	"time"

	"github.com/oklog/ulid/v2"
)

func LoadResource(r *ResourceInfo) error {
	if r.Status != RState_ReadyToProcess {
		return errors.New("resource is not ready to be processed")
	}

	dirPath := filepath.Dir(r.ResourcePath)
	if _, err := os.Stat(dirPath); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(dirPath, os.ModePerm)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	if _, err := os.Stat(r.RawFilePath); err != nil {
		//Deveria fazer algo a respeito desse file path que não existe
		//Ou removê-lo completamente ou colocar num status de erro e notificar o content manager
		fmt.Printf("Ingest file was not found on path %v\n", r.RawFilePath)
		return err
	}

	fmt.Println("Updating resource status to processing")
	r.Status = RState_Processing
	err := UpdateResource(r)
	if err != nil {
		fmt.Printf("an error occured while trying to update resource %v\n", r.id)
		return err
	}

	fmt.Println("Converting ingest media to HLS")
	fmt.Printf("Converting %v to %v", r.RawFilePath, r.ResourcePath)
	//ffmpeg -i sample.mkv -c:a copy -f hls -hls_playlist_type vod output.m3u8
	cmd := exec.Command("ffmpeg",
		"-i", r.RawFilePath,
		"-c:a", "copy",
		"-f", "hls",
		"-hls_playlist_type", "vod",
		r.ResourcePath,
	)

	stdout, err := cmd.Output()

	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	r.LoadedDate = time.Now()
	r.Status = RState_Loaded

	// Print the output
	fmt.Println(string(stdout))
	fmt.Printf("Conversão finalizada às %v", r.LoadedDate)

	UpdateResource(r)

	return nil
}

func UpdateResource(r *ResourceInfo) error {

	loadDate := r.LoadedDate.Local().Format(time.RFC3339)
	createDate := r.CreatedDate.Local().Format(time.RFC3339)

	return dtaccess.UpdateResource(&dtaccess.ResourceData{
		Id:                 r.id,
		Resource_id:        r.ResourceId,
		Resource_path:      r.ResourcePath,
		Manifest_file_name: r.ManifestFileName,
		Raw_file_path:      r.RawFilePath,
		Raw_file_name:      r.RawFileName,
		Loaded_date:        loadDate,
		Created_date:       createDate,
		Resource_status:    r.Status,
	})
}

func CreateResource(ingestPath string) (*ResourceInfo, error) {

	fmt.Println("Executando CreateResource")

	fileInfo, err := os.Stat(ingestPath)
	if err != nil {
		return nil, err
	}

	fullpath, err := filepath.Abs(ingestPath)
	if err != nil {
		return nil, err
	}

	r := ResourceInfo{
		RawFilePath: fullpath,
		RawFileName: fileInfo.Name(),
		Status:      RState_New,
	}

	if !isAcceptedVideoExtension(r.GetIngestFileExtension()) {
		return nil, errors.New("unsupported file extension")
	}

	fmt.Printf("ingest fileName: %v\n", r.RawFileName)
	fmt.Printf("ingest filePath: %v\n", r.RawFilePath)
	fmt.Printf("ingest fileExt: %v\n", r.GetIngestFileExtension())

	r.ResourceId = ulid.Make().String()

	r.ManifestFileName = r.ResourceId + "_manifest.m3u8"

	rpf, err := appconfigs.GetConfig(appconfigs.RootPackagesFolder)
	if err != nil {
		//Uses default path
		rpf = "./resources/hls/"
	}

	rpfInfo, err := os.Stat(rpf)
	if err != nil || !rpfInfo.IsDir() {
		return nil, errors.New("unable to find root packages dir")
	}
	r.ResourcePath = filepath.Join(rpf, r.ResourceId, r.ManifestFileName)

	fmt.Printf("ManifestFileName: %v\n", r.ManifestFileName)
	fmt.Printf("ResourcePath: %v\n", r.ResourcePath)

	resourceData := dtaccess.ResourceData{
		Resource_id:        r.ResourceId,
		Resource_path:      r.ResourcePath,
		Manifest_file_name: r.ManifestFileName,
		Raw_file_path:      r.RawFilePath,
		Raw_file_name:      r.RawFileName,
		Resource_status:    r.Status,
	}

	err = dtaccess.CreateResource(&resourceData)
	if err != nil {
		return nil, err
	}

	fmt.Println("Resource created succesfully")
	r.id = resourceData.Id

	return &r, nil
}

func resourceDataToResourceInfo(rdata *dtaccess.ResourceData) (*ResourceInfo, error) {
	loadDate, err := time.Parse(time.RFC3339, rdata.Loaded_date)
	if err != nil {
		loadDate = time.Time{}
	}

	createDate, err := time.Parse(time.RFC3339, rdata.Created_date)
	if err != nil {
		createDate = time.Time{}
	}

	rInfo := ResourceInfo{
		id:               rdata.Id,
		ResourceId:       rdata.Resource_id,
		ResourcePath:     rdata.Resource_path,
		ManifestFileName: rdata.Manifest_file_name,
		RawFilePath:      rdata.Raw_file_path,
		RawFileName:      rdata.Raw_file_name,
		LoadedDate:       loadDate,
		CreatedDate:      createDate,
		Status:           rdata.Resource_status,
	}

	return &rInfo, nil
}

func GetResourceInfoById(id string) (*ResourceInfo, error) {
	rdata, err := dtaccess.GetResourceById(id)
	if err != nil {
		if errors.Is(err, &dtaccess.DbNotFound{}) {
			//maybe log it
			return nil, errors.New("resource not found")
		}

		return nil, err
	}
	return resourceDataToResourceInfo(rdata)
}

func GetResourcesToLoad() ([]ResourceInfo, error) {
	rdataSlices, err := dtaccess.GetResourcesToLoad()
	if err != nil {
		return nil, err
	}

	resourcesSlice := make([]ResourceInfo, len(rdataSlices))
	for i, r := range rdataSlices {
		loadDate, err := time.Parse(time.RFC3339, r.Loaded_date)
		if err != nil {
			loadDate = time.Time{}
		}

		createDate, err := time.Parse(time.RFC3339, r.Created_date)
		if err != nil {
			createDate = time.Time{}
		}

		resourcesSlice[i] = ResourceInfo{
			id:               r.Id,
			ResourceId:       r.Resource_id,
			ResourcePath:     r.Resource_path,
			ManifestFileName: r.Manifest_file_name,
			RawFilePath:      r.Raw_file_path,
			RawFileName:      r.Raw_file_name,
			LoadedDate:       loadDate,
			CreatedDate:      createDate,
			Status:           r.Resource_status,
		}
	}

	return resourcesSlice, nil
}

func GetResourceInfoByFileName(fileName string) (*ResourceInfo, error) {
	rdata, err := dtaccess.GetResourceByFileName(fileName)
	if err != nil {
		if errors.Is(err, &dtaccess.DbNotFound{}) {
			//maybe log it
			return nil, nil
		}

		return nil, err
	}
	return resourceDataToResourceInfo(rdata)
}

func CreateMovie(movie *Movie) (*Movie, error) {
	mData := dtaccess.MovieData{
		Resource_id: movie.ResourceId,
		File_name:   movie.FileName,
		Title:       movie.Title,
		Year:        movie.Year,
		Rated:       movie.Rated,
		Released:    movie.Released,
		Runtime:     movie.Runtime,
		Genre:       movie.Genre,
		Director:    movie.Director,
		Writer:      movie.Writer,
		Actors:      movie.Actors,
		Plot:        movie.Plot,
		Language:    movie.Language,
		Country:     movie.Country,
		Awards:      movie.Awards,
		Poster:      movie.Poster,
		Metascore:   movie.Metascore,
		Imdb_rating: movie.ImdbRating,
		Imdb_votes:  movie.ImdbVotes,
		Imdb_id:     movie.ImdbID,
		Type:        movie.Type,
		DVD:         movie.DVD,
		Production:  movie.Production,
		Website:     movie.Website,
	}

	err := dtaccess.CreateMovie(&mData)
	if err != nil {
		return nil, err
	}

	movie.id = mData.Id

	return movie, nil
}

func GetMovieByFileName(fileName string) (*Movie, error) {
	mData, err := dtaccess.GetMovieByFileName(fileName)
	if err != nil {
		if errors.Is(err, &dtaccess.DbNotFound{}) {
			return nil, nil
		}
		return nil, err
	}

	return &Movie{
		id:         mData.Id,
		ResourceId: mData.Resource_id,
		FileName:   mData.File_name,
		Title:      mData.Title,
		Year:       mData.Year,
		Rated:      mData.Rated,
		Released:   mData.Released,
		Runtime:    mData.Runtime,
		Genre:      mData.Genre,
		Director:   mData.Director,
		Writer:     mData.Writer,
		Actors:     mData.Actors,
		Plot:       mData.Plot,
		Language:   mData.Language,
		Country:    mData.Country,
		Awards:     mData.Awards,
		Poster:     mData.Poster,
		Metascore:  mData.Metascore,
		ImdbRating: mData.Imdb_rating,
		ImdbVotes:  mData.Imdb_votes,
		ImdbID:     mData.Imdb_id,
		Type:       mData.Type,
		DVD:        mData.DVD,
		Production: mData.Production,
		Website:    mData.Website,
	}, nil
}
