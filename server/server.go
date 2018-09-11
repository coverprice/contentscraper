package server

import (
	"fmt"
	"github.com/coverprice/contentscraper/drivers"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"path/filepath"
)

// The web server that displays the content scraped by the harvesting drivers.
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

	// Add the static directory so Javascript can be served
	prefix := "/static/"
	exec_path, err := os.Executable()
	if err != nil {
		log.Fatal("Could not detect directory of executable", err)
	}
	static_dir := filepath.Join(filepath.Dir(exec_path), prefix)
	mux.Handle(prefix, http.StripPrefix(prefix, http.FileServer(http.Dir(static_dir))))

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
