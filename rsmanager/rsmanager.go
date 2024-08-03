package rsmanager

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type ResourceInfo struct {
	ResourceId, ResourcePath, ManifestFileName, RawFilePath, RawFileName string
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

	// Print the output
	fmt.Println(string(stdout))
	fmt.Println("Conversão finalizada")

	return nil
}

func GetResourceInfoById(id string) (ResourceInfo, error) {
	for _, r := range resources {
		if r.ResourceId == id {
			return r, nil
		}
	}

	return ResourceInfo{}, errors.New("resource not found")
}
