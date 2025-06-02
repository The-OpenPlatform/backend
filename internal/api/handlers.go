package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/jmoiron/sqlx"
	"net/http"
	"os"
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Welcome to the root endpoint!"))
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, World!"))
}

func getStatus(w http.ResponseWriter, r *http.Request) {
	podName, err := os.Hostname()
	if err != nil {
		podName = "unknown"
	}
	w.Write([]byte(podName))
}

func GetModulesWithImages(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var modules []Module
		err := db.Select(&modules, `SELECT module_id, name FROM modules`)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for i := range modules {
			var images []struct {
				ModuleID   string `db:"module_id"`
				Image      []byte `db:"image"`
				FileFormat string `db:"fileformat"`
			}

			err := db.Select(&images, `SELECT module_id, image, fileformat FROM images WHERE module_id=$1`, modules[i].ModuleID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			for _, img := range images {
				base64Data := base64.StdEncoding.EncodeToString(img.Image)
				dataURL := fmt.Sprintf("data:%s;base64,%s", img.FileFormat, base64Data)
				modules[i].Images = append(modules[i].Images, Image{ModuleID: img.ModuleID, DataURL: dataURL})
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(modules)
	}
}

type Module struct {
	ModuleID string  `db:"module_id" json:"module_id"`
	Name     string  `db:"name" json:"name"`
	Images   []Image `json:"images"`
}

type Image struct {
	ModuleID string `json:"module_id"`
	DataURL  string `json:"data_url"` // base64 image data
}
