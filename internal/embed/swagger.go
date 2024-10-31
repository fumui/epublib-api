package embedServer

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gorilla/mux"
)

func RegisterSwaggerUI(embed embed.FS, dirName string, mux *mux.Router) {
	subFS, _ := fs.Sub(embed, dirName)
	distFs := http.FileServer(http.FS(subFS))

	mux.PathPrefix("/api/v1/swagger").Handler(http.StripPrefix("/api/v1/swagger", distFs))
}
