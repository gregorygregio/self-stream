package rsmanager

import (
	"os"
	"strings"
	"time"
)

type ResourceInfo struct {
	id                                                                   int32
	ResourceId, ResourcePath, ManifestFileName, RawFilePath, RawFileName string
	LoadedDate, CreatedDate                                              time.Time
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
