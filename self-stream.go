package main

import (

	//"net/http"
	"fmt"
	"log"
	"net/http"
	"self-stream/appconfigs"
	"self-stream/dtaccess"
	"self-stream/rsmanager"
	vidstreaming "self-stream/videostreaming"
	"strings"
	"time"
	//"strings"
	//"time"
	//"github.com/gin-gonic/gin"
)

func main() {

	dtaccess.InitDb()

	appconfigs.LoadConfigs()

	go rsmanager.StartResourceWorker()

	port := 8080

	hlsServce := vidstreaming.HlsServer{
		Route:       "/",
		ContentPath: "./resources/hls",
	}

	hlsServce.StartHlsServer()

	http.Handle("/home", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		r := strings.NewReader(home)
		http.ServeContent(rw, req, "index.html", time.Time{}, r)
	}))

	//router := gin.Default()

	//router.GET("/resources/:resourceId", getResourceById)

	fmt.Printf("Starting HLS server on %v\n", port)
	log.Printf("Serving %s on HTTP port: %v\n", hlsServce.ContentPath, port)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), nil))

}

const home = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>HLS demo</title>
<script src="https://cdn.jsdelivr.net/npm/hls.js@latest"></script>
</head>
<body>
<video id="video" muted autoplay controls height="450"></video>
<script>
let hls = new Hls();
hls.loadSource('http://localhost:8080/balcony/output.m3u8');
hls.attachMedia(document.getElementById('video'));
hls.on(Hls.Events.MANIFEST_PARSED, () => video.play());
</script>
</body>
</html>
`
