package main

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var htmlTemplate = `
<!DOCTYPE html>
<html>
	<head>
		<title>Image Processor</title>
	</head>
	<body>
		<form action="/process" method="post">
			<h3>Enter image URLs:</h3>
			<input type="text" name="url1" size="100"><br>
			<input type="text" name="url2" size="100"><br>
			<input type="text" name="url3" size="100"><br>
			<input type="text" name="url4" size="100"><br>
			<input type="text" name="url5" size="100"><br>
			<label for="color">Hex Color Code:</label><br>
			<input type="color" name="color" id="color" value="#FF0000">
			<input type="submit" value="Submit">
		</form>
		{{if .}}
			<h2>Processed Images</h2>
			{{range .}}
				<img src="{{.}}" width="300">
			{{end}}
		{{end}}
	</body>
</html>
`

var tpl *template.Template

func init() {
	tpl = template.Must(template.New("webpage").Parse(htmlTemplate))
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

func processImages(w http.ResponseWriter, r *http.Request) {
	//urls := strings.Split(r.FormValue("urls"), ",")
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

		fmt.Println("Image Saved " + filePath)

		newFilePath, err := ProcessImage(filePath, parseHexColor(color))
		if err != nil {
			http.Error(w, "Error processing image: "+err.Error(), http.StatusInternalServerError)
		}

		processedImagePaths = append(processedImagePaths, newFilePath)
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
