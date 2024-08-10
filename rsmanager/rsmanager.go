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
	"time"
)

func LoadResource(r *ResourceInfo) error {
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
	}

	return &rInfo, nil
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
		}
	}

	return resourcesSlice, nil
}
