package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fogleman/gg"
)

type Config struct {
	Scale       float64 `json:"scale"`
	BorderWidth float64 `json:"borderWidth"`
	OutputDir   string  `json:"outputDir"`
}

func main() {
	path, err := ExecDir("config.json")
	if err != nil {
		fmt.Println("Error reading config:", err)
		fmt.Scanf("h")
		os.Exit(1)
	}
	config, err := readConfig(path)
	if err != nil {
		fmt.Println("Error reading config:", err)
		fmt.Scanf("h")
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		fmt.Println("Drag and drop image files onto this executable.")
		fmt.Scanf("h")
		os.Exit(1)
	}

	if config.OutputDir[len(config.OutputDir)-1] != '/' {
		config.OutputDir = config.OutputDir + "/"
	}

	err = createOutputDir(config.OutputDir)
	if err != nil {
		fmt.Println("Unable to create output directory")
		fmt.Scanf("h")
		os.Exit(1)
	}

	for _, arg := range os.Args[1:] {
		_, err := os.Stat(arg)
		if os.IsNotExist(err) {
			fmt.Printf("File not found: %s\n", arg)
			fmt.Scanf("h")
			continue
		}

		if isImageFile(arg) {
			err = handleFile(arg, &config)
			if err != nil {
				fmt.Println("Error:", err)
				fmt.Scanf("h")
			} else {
				fmt.Printf("Image processed %s\n", arg)
			}
		} else {
			fmt.Printf("Not an image file: %s\n", arg)
		}
	}

	exitChan := make(chan struct{})

	// Set the duration to 5 seconds
	duration := 5 * time.Second

	fmt.Printf("Finished, exiting in %s...", duration.String())
	// Start a goroutine to wait for the specified duration
	go func() {
		time.Sleep(duration)
		close(exitChan) // Signal that the time is up
	}()

	// Wait for the signal or until the time is up
	<-exitChan

}

func createOutputDir(path string) error {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			return err
		}
	}

	return nil
}

func wrap(ft string, args ...any) error {
	s := fmt.Sprintf(ft, args...)
	return errors.New(s)
}

func handleFile(file string, config *Config) error {
	img, err := gg.LoadImage(file)
	if err != nil {
		return wrap("Error loading image %s: %s\n", file, err)
	}

	img, err = processImage(img, config)
	if err != nil {
		return wrap("Error processing image %s: %s\n", file, err)
	}

	newFilename := config.OutputDir + "token_" + filepath.Base(strings.Replace(file, "danimalsound_", "", 1))
	err = gg.SavePNG(newFilename, img)
	if err != nil {
		return wrap("Error saving image %s: %s\n", newFilename, err)
	}

	return nil
}

// Check if a file has an image extension (you can expand this list)
func isImageFile(filePath string) bool {
	ext := filepath.Ext(filePath)
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp":
		return true
	default:
		return false
	}
}

func readConfig(filePath string) (Config, error) {
	var config Config

	file, err := os.Open(filePath)
	if err != nil {
		return config, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return config, err
	}

	return config, nil
}

func ExecDir(relPath string) (string, error) {
	// os.Executable requires Go 1.18+
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}

	return filepath.Join(filepath.Dir(ex), relPath), nil
}
