package rsmanager

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"self-stream/appconfigs"
	"self-stream/dtaccess"
	"strings"
	"time"
)

type ResourceInfo struct {
	id                                                                   int32
	ResourceId, ResourcePath, ManifestFileName, RawFilePath, RawFileName string
	LoadedDate, CreatedDate                                              time.Time
}

var resources = []ResourceInfo{
	{
		ResourceId:       "wdad15wd1a31",
		ResourcePath:     "resources/hls/balcony_wdad15wd1a31",
		ManifestFileName: "balcony_wdad15wd1a31.m3u8",
		RawFilePath:      "resources/raw",
		RawFileName:      "balcony.mp4",
	},
}

func (r *ResourceInfo) IsResourceLoaded() bool {
	_, err := os.Stat(r.ResourcePath + "/" + r.ManifestFileName)
	return err == nil
}

func (r *ResourceInfo) GetIngestFileExtension() string {
	fnameSlice := strings.Split(r.RawFileName, ".")
	if len(fnameSlice) > 0 {
		return "." + fnameSlice[len(fnameSlice)-1]
	}
	return ""
}

func (r *ResourceInfo) LoadResource() error {
	if _, err := os.Stat(r.ResourcePath); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(r.ResourcePath, os.ModePerm)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	fmt.Println("Iniciando conversão de vídeo raw para HLS")

	wdpath, err := os.Getwd()
	if err != nil {
		return err
	}
	rawSourcePath := filepath.Join(wdpath, r.RawFilePath+"/"+r.RawFileName)
	destPath := filepath.Join(wdpath, r.ResourcePath+"/"+r.ManifestFileName)

	fmt.Printf("Convertendo %v para %v", rawSourcePath, destPath)
	//ffmpeg -i sample.mkv -c:a copy -f hls -hls_playlist_type vod output.m3u8
	cmd := exec.Command("ffmpeg",
		"-i", rawSourcePath,
		"-c:a", "copy",
		"-f", "hls",
		"-hls_playlist_type", "vod",
		destPath,
	)

	stdout, err := cmd.Output()

	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	r.LoadedDate = time.Now()

	// Print the output
	fmt.Println(string(stdout))
	fmt.Printf("Conversão finalizada às %v", r.LoadedDate)

	r.UpdateResource()

	return nil
}

func (r *ResourceInfo) UpdateResource() error {

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
		//ResourcePath:     "resources/hls/balcony_wdad15wd1a31",
		//ManifestFileName: "balcony_wdad15wd1a31.m3u8",
		RawFilePath: fullpath,
		RawFileName: fileInfo.Name(),
	}

	if !isAcceptedVideoExtension(r.GetIngestFileExtension()) {
		return nil, errors.New("unsupported file extension")
	}

	fmt.Printf("ingest fileName: %v\n", r.RawFileName)
	fmt.Printf("ingest filePath: %v\n", r.RawFilePath)
	fmt.Printf("ingest fileExt: %v\n", r.GetIngestFileExtension())

	h := sha256.New()
	h.Write([]byte(r.RawFileName))
	hash := h.Sum(nil)
	fmt.Printf("Criando resource_id %x\n", hash)

	r.ResourceId = hex.EncodeToString(hash)

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
	}

	err = dtaccess.CreateResource(&resourceData)
	if err != nil {
		return nil, err
	}

	fmt.Println("Resource created succesfully")
	r.id = resourceData.Id

	return &r, nil
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

	loadDate, err := time.Parse(time.RFC3339, rdata.Loaded_date)
	if err != nil {
		//logError
	}

	createDate, err := time.Parse(time.RFC3339, rdata.Created_date)
	if err != nil {
		//logError
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
	}

	return &rInfo, nil
}

var acceptedVideoExtensions = []string{
	"mp4",
	"wav",
	"mkv",
	"ico",
}

func isAcceptedVideoExtension(e string) bool {
	e = strings.Replace(e, ".", "", 1)
	for _, ext := range acceptedVideoExtensions {
		if e == ext {
			return true
		}
	}

	return false
}
