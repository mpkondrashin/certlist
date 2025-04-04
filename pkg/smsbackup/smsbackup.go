package smsbackup

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

var ErrNotFound = errors.New("file not found in tag.gz archive")

func ExtractDump(backupPath string) (string, error) {
	folder := filepath.Dir(backupPath)
	fileName := "noalerts.mysqldump"
	filePath := filepath.Join(folder, fileName)
	return filePath, extractFileFromZIP(backupPath, fileName, filePath)
}

/*
	func extractFileFromTarGz(tarGzPath, targetFile, outputPath string) error {
		file, err := os.Open(tarGzPath)
		if err != nil {
			return err
		}
		defer file.Close()

		gzReader, err := gzip.NewReader(file)
		if err != nil {
			return err
		}
		defer gzReader.Close()

		tarReader := tar.NewReader(gzReader)
		for {
			header, err := tarReader.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}
			if header.Name == targetFile {
				outFile, err := os.Create(outputPath)
				if err != nil {
					return err
				}
				defer outFile.Close()

				_, err = io.Copy(outFile, tarReader)
				if err != nil {
					return err
				}

				fmt.Printf("Extracted %s to %s\n", targetFile, outputPath)
				return nil
			}
		}
		return fmt.Errorf("%s: %w", targetFile, ErrNotFound)
	}
*/
func extractFileFromZIP(zipPath, targetFile, outputPath string) error {
	zipReader, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer zipReader.Close()
	for _, file := range zipReader.File {
		if file.Name != targetFile {
			continue
		}
		rc, err := file.Open()
		if err != nil {
			return err
		}
		defer rc.Close()
		outFile, err := os.Create(outputPath)
		if err != nil {
			return err
		}
		defer outFile.Close()
		_, err = io.Copy(outFile, rc)
		if err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("%s: %w", targetFile, ErrNotFound)
}
