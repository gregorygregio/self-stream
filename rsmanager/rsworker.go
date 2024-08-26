package rsmanager

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

const (
	MediaType_Movie  = 1
	MediaType_Series = 2
)

func StartResourceWorker() {
	fmt.Println("Starting Resources worker")

	workerTicker := time.NewTicker(30 * time.Second)

	onWorkerTick()
	for range workerTicker.C {
		onWorkerTick()
	}
}

func onWorkerTick() {
	go searchMovies()
	//go searchSeries()
	// fmt.Println("Tick")
	// resourcesToLoad, err := GetResourcesToLoad()
	// if err != nil {
	// 	fmt.Printf("Error: %v\n", err.Error())
	// }

	// fmt.Printf("Number of resources to Load: %v\n", len(resourcesToLoad))

	// for i := 0; i < len(resourcesToLoad); i++ {
	// 	go processResource(&resourcesToLoad[i])
	// }
}

func fetchMediaFilesIteratively(dirPath string, subPath string) ([]string, error) {
	mediaFilesList := make([]string, 0)
	if _, err := os.Stat(dirPath); errors.Is(err, os.ErrNotExist) {
		fmt.Printf("Media folder %v does not exist or is not accessible.\n", dirPath)
		return nil, err
	}

	dirEntries, err := os.ReadDir(dirPath)
	if err != nil {
		fmt.Printf("Could not list content of folder %v.\n", dirPath)
		fmt.Printf("Could not list content of folder %v.\n", err.Error())
		return nil, err
	}

	for _, x := range dirEntries {
		if x.IsDir() {
			subDirFiles, err := fetchMediaFilesIteratively(filepath.Join(dirPath, x.Name()), x.Name())
			if err != nil {
				return nil, err
			}

			mediaFilesList = append(mediaFilesList, subDirFiles...)
			continue
		}

		fnameSplit := strings.Split(x.Name(), ".")
		if len(fnameSplit) < 2 {
			continue
		}
		extension := fnameSplit[len(fnameSplit)-1]
		if isAcceptedVideoExtension(extension) {
			//mediaFilesList = append(mediaFilesList, x.Name())
			//mediaFilesList = append(mediaFilesList, filepath.Join(dirPath, x.Name()))
			mediaFilesList = append(mediaFilesList, filepath.Join(subPath, x.Name()))
		}
	}

	return mediaFilesList, nil
}

var movieFilesHash = ""

const moviesRootFolder = "/path/to/Movies"

func checkIfMoviesFolderChanged(movieMediaFiles []string) bool {
	h := sha256.New()
	for _, mf := range movieMediaFiles {
		h.Write([]byte(mf))
	}
	hash := h.Sum(nil)
	if movieFilesHash == hex.EncodeToString(hash) {
		fmt.Printf("Hash: %v\n", hex.EncodeToString(hash))
		return false
	}
	movieFilesHash = hex.EncodeToString(hash)
	return true
}

func searchMovies() {
	fmt.Println("Starting search for movies")

	movieMediaFiles, err := fetchMediaFilesIteratively(moviesRootFolder, "")
	if err != nil {
		fmt.Printf("There was an error fetching movies: %v\n", err.Error())
	}

	if !checkIfMoviesFolderChanged(movieMediaFiles) {
		fmt.Println("No changes on the movies folder")
		return
	}

	fmt.Println("The movies folder changed!")

	for _, mediaFilePath := range movieMediaFiles {
		createMovieIfNotExists(mediaFilePath)
	}

	fmt.Println("Finished search for movies")
}

func createMovieIfNotExists(mediaFilePath string) {
	_, mediaFileName := path.Split(mediaFilePath)

	rInfo, err := GetResourceInfoByFileName(mediaFileName)

	if err != nil {
		fmt.Printf("There was an error to fetch resource %v\n", mediaFilePath)
		fmt.Println(err.Error())
		return
	}

	if rInfo == nil {
		rInfo, err = CreateResource(filepath.Join(moviesRootFolder, mediaFilePath))
		if err != nil {
			fmt.Printf("There was an error to create resource %v\n", mediaFilePath)
			fmt.Println(err.Error())
			return
		}
	}
	movie := Movie{}
	movie.ResourceId = rInfo.ResourceId
	movie.FileName = mediaFileName
	movie.Title = mediaFileName

	existingMovie, err := GetMovieByFileName(movie.FileName)
	if err != nil {
		fmt.Printf("There was an error to fetch movie %v\n", movie.FileName)
		fmt.Println(err.Error())
		return
	}
	if existingMovie == nil {
		CreateMovie(&movie)
	}
}

func searchSeries() {
	fmt.Println("Starting search for tv series")
	seriesRootFolder := "/path/to/Series"

	movieMediaFiles, err := fetchMediaFilesIteratively(seriesRootFolder, "")
	if err != nil {
		fmt.Printf("There was an error fetching tv series: %v\n", err.Error())
	}
	for _, mf := range movieMediaFiles {
		fmt.Println(mf)
	}

	fmt.Println("Finished search for tv series")
}

func processResource(resource *ResourceInfo) {
	fmt.Printf("Working on resource %v\n", resource.RawFileName)
	err := LoadResource(resource)
	if err != nil {
		fmt.Printf("Error to load resource: %v\n", err.Error())
	}
}
