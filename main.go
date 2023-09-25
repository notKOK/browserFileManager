package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var (
	currentPath string
	initialPage = "index.html"
	renamePage  = "rename.html"
	workDir     = "/workDir"
)

type FileInfo struct {
	Name  string
	IsDir bool
}

func main() {
	http.HandleFunc("/", listHandler)
	http.HandleFunc("/create", createHandler)
	http.HandleFunc("/delete", deleteHandler)
	http.HandleFunc("/rename", renameHandler)
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/download", downloadHandler)
	http.HandleFunc("/renamed", renamedHandler)

	currentPath, _ = os.Getwd()
	currentPath += workDir

	_, err := os.Stat(currentPath)

	if os.IsNotExist(err) {
		err := os.MkdirAll(currentPath, 0755)
		if err != nil {
			log.Fatal(err)
			return
		}
	} else if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Server started at http://localhost:8080\n")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		return
	}
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	files, err := listFiles(currentPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmplData := struct {
		Files    []FileInfo
		CurrPath string
	}{
		Files:    files,
		CurrPath: currentPath,
	}

	var t *template.Template

	t, err = template.ParseFiles(initialPage)
	err = t.Execute(w, tmplData)
	if err != nil {
		return
	}
}

func listFiles(dirPath string) ([]FileInfo, error) {
	var files []FileInfo

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		files = append(files, FileInfo{Name: info.Name(), IsDir: info.IsDir()})
		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		newName := r.FormValue("newName")
		newPath := filepath.Join(currentPath, newName)

		if !strings.HasSuffix(newPath, "/") {
			newPath += "/"
		}

		err := os.MkdirAll(newPath, os.ModePerm)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	path := filepath.Join(currentPath, name)

	err := os.Remove(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func renameHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")

	t, err := template.ParseFiles(renamePage)
	err = t.Execute(w, name)
	if err != nil {
		return
	}

	return

}

func renamedHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		oldName := r.FormValue("oldName")
		newName := r.FormValue("newName")

		oldPath := filepath.Join(currentPath, oldName)
		newPath := filepath.Join(currentPath, newName)

		err := os.Rename(oldPath, newPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()

		uploadPath := filepath.Join(currentPath, header.Filename)

		newFile, err := os.Create(uploadPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer newFile.Close()

		_, err = io.Copy(newFile, file)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	filePath := filepath.Join(currentPath, name)

	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", name))

	_, err = io.Copy(w, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
