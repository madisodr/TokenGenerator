package main

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var tpl *template.Template

func init() {
	tpl = template.Must(template.ParseFiles("index.html"))
}

func main() {
	http.HandleFunc("/", renderForm)
	http.HandleFunc("/process", processImages)
	http.HandleFunc("/images/", serveImage)
	http.ListenAndServe(":8080", nil)
}

func renderForm(w http.ResponseWriter, r *http.Request) {
	tpl.Execute(w, nil)
}

func handleFileUpload(r *http.Request) ([]string, error) {
	color := r.FormValue("color")
	processedImagePaths := make([]string, 0)

	for i := 1; i <= 5; i++ {
		file, _, err := r.FormFile(fmt.Sprintf("image%d", i))
		if err != nil {
			if err == http.ErrMissingFile {
				// Skip if no file was uploaded in this input
				continue
			}
			fmt.Println("Error retrieving the file:", err)
			return processedImagePaths, err
		}
		defer file.Close()

		// Create a temporary file to save the uploaded file
		tempFile, err := ioutil.TempFile("images", "token_*.png")
		if err != nil {
			fmt.Println("Error creating temp file:", err)
			return processedImagePaths, err
		}
		defer tempFile.Close()

		// Save the uploaded file to the temporary file
		_, err = io.Copy(tempFile, file)
		if err != nil {
			fmt.Println("Error saving file:", err)
			return processedImagePaths, err
		}

		fmt.Println("Uploaded File:", tempFile.Name())

		// Process the image and add to the list
		newFilePath, err := ProcessImage(tempFile.Name(), parseHexColor(color))
		if err != nil {
			return processedImagePaths, err
		}

		processedImagePaths = append(processedImagePaths, newFilePath)
	}

	return processedImagePaths, nil
}

func handleUrlUpload(r *http.Request) ([]string, error) {
	color := r.FormValue("color")
	processedImagePaths := make([]string, 0)

	for i := 1; i <= 5; i++ {
		url := r.FormValue(fmt.Sprintf("url%d", i))
		fmt.Println(url)
		if url == "" {
			continue
		}

		url = trimAfterExtension(strings.TrimSpace(url))
		filePath, err := downloadImage(url)

		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		fmt.Println("Uploaded File:", filePath)

		newFilePath, err := ProcessImage(filePath, parseHexColor(color))
		if err != nil {
			return processedImagePaths, err
		}

		processedImagePaths = append(processedImagePaths, newFilePath)
	}

	return processedImagePaths, nil
}

func processImages(w http.ResponseWriter, r *http.Request) {
	isFileUpload := r.FormValue("toggle") == "on"
	var processedImagePaths []string
	var err error

	if isFileUpload {
		processedImagePaths, err = handleFileUpload(r)
		if err != nil {
			http.Error(w, "Error processing image: "+err.Error(), http.StatusInternalServerError)
		}
	} else {
		processedImagePaths, err = handleUrlUpload(r)
		if err != nil {
			http.Error(w, "Error processing image: "+err.Error(), http.StatusInternalServerError)
		}
	}

	tpl.Execute(w, processedImagePaths)
}

func trimAfterExtension(inputURL string) string {
	re := regexp.MustCompile(`(\.png|\.jpeg)(\?.*|$)`) // Regex to match .png or .jpeg followed by a query string or end of string
	return re.ReplaceAllString(inputURL, "$1")
}

func serveImage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, r.URL.Path[1:])
}

func downloadImage(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	tokens := strings.Split(url, "/")
	fileName := "images/" + tokens[len(tokens)-1]

	ext := filepath.Ext(fileName)
	if ext != ".png" && ext != ".jpeg" {
		return "", errors.New("file is not a png or jpeg")
	}

	os.MkdirAll("images", os.ModePerm)
	file, err := os.Create(fileName)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return fileName, err
}
