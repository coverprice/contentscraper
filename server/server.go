package server

import (
	"fmt"
	"github.com/coverprice/contentscraper/drivers"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type Server struct {
	server  http.Server
	mux     *http.ServeMux
	Drivers []drivers.IDriver
}

func NewServer(port int) *Server {
	mux := http.NewServeMux()
	s := Server{
		server: http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: mux,
		},
		mux: mux,
	}
	mux.Handle("/", indexHandler{server: &s})
	return &s
}

func (this *Server) AddDriver(driver drivers.IDriver) {
	this.Drivers = append(this.Drivers, driver)

	var baseUrlPath = driver.GetBaseUrlPath()
	this.mux.Handle(
		baseUrlPath,
		http.StripPrefix(baseUrlPath, driver.GetHttpHandler()),
	)
}

// Does not return!
func (this *Server) Launch() {
	err := this.server.ListenAndServe()
	if err != nil {
		log.Fatal("Failed to launch web server: ", err)
	}
}
