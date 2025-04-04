package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/tidwall/gjson"
)

const (
	version        = "11.7"
	outputFilename = "mariadb-latest.zip"
)

type Download struct {
	DownloadURL string `json:"download_url"`
	Platform    string `json:"platform"`
	Arch        string `json:"arch"`
	Extension   string `json:"extension"`
}

type Release struct {
	Downloads []Download `json:"downloads"`
}

func main() {
	log.Println("Fetching latest release info from MariaDB API...")
	apiURL := fmt.Sprintf("https://downloads.mariadb.org/rest-api/mariadb/%s/latest/", version)
	res, err := http.Get(apiURL)
	if err != nil {
		log.Fatalf("Error fetching data from API: %v\n", err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Fatalf("Failed to fetch release info. HTTP Status: %d\n", res.StatusCode)
	}

	var responseBody bytes.Buffer
	_, err = io.Copy(&responseBody, res.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v\n", err)
	}

	files := gjson.GetBytes(responseBody.Bytes(), "releases.*.files")

	downloadURL := ""
	sha256hash := ""
	files.ForEach(func(_, file gjson.Result) bool {
		if file.Get("os").String() != "Windows" {
			return true
		}
		if file.Get("cpu").String() != "x86_64" {
			return true
		}
		if file.Get("package_type").String() != "ZIP file" {
			return true
		}
		if strings.Contains(file.Get("file_name").String(), "debugsymbols") {
			return true
		}
		downloadURL = file.Get("file_download_url").String()
		sha256hash = file.Get("checksum.sha256sum").String()
		return true
	})
	if downloadURL == "" {
		log.Fatal("No suitable download URL found.")
	}
	fmt.Printf("Downloading MariaDB ZIP from: %s\n", downloadURL)

	resp, err := http.Get(downloadURL)
	if err != nil {
		log.Fatalf("Error downloading file: %v\n", err)
	}
	defer resp.Body.Close()

	out, err := os.Create(outputFilename)
	if err != nil {
		log.Fatalf("Error creating file: %v\n", err)
	}
	defer out.Close()

	hash := sha256.New()
	multiWriter := io.MultiWriter(out, hash)

	_, err = io.Copy(multiWriter, resp.Body)
	if err != nil {
		log.Fatalf("Error saving file: %v\n", err)
	}

	hashResult := hash.Sum(nil)
	if fmt.Sprintf("%x", hashResult) != sha256hash {
		log.Fatalf("SHA256 hash mismatch: expected %s, got %x\n", sha256hash, hashResult)
	} else {
		fmt.Println("SHA256 hash verified successfully.")
	}
	fmt.Printf("Download completed: %s\n", outputFilename)
}
