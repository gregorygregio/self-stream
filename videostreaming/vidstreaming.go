package vidstreaming

import (
	"net/http"
)

type HlsServer struct {
	Route, ContentPath string
}

func (s *HlsServer) StartHlsServer() {

	http.Handle(s.Route, http.FileServer(http.Dir(s.ContentPath)))
}
