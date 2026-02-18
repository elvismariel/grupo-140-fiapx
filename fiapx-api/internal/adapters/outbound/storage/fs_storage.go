package storage

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"video-processor/internal/core/domain"
	"video-processor/internal/core/ports"
)

type fsStorage struct {
	uploadDir string
	outputDir string
	tempDir   string
}

func NewFSStorage() ports.Storage {
	storage := &fsStorage{
		uploadDir: "/app/uploads",
		outputDir: "/app/outputs",
		tempDir:   "/app/temp",
	}
	storage.createDirs()
	return storage
}

func (s *fsStorage) createDirs() {
	dirs := []string{s.uploadDir, s.outputDir, s.tempDir}
	for _, dir := range dirs {
		os.MkdirAll(dir, 0755)
	}
}

func (s *fsStorage) SaveUpload(filename string, data io.Reader) (string, error) {
	path := filepath.Join(s.uploadDir, filename)
	out, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, data)
	return path, err
}

func (s *fsStorage) SaveZip(zipFilename string, files []string) error {
	zipPath := filepath.Join(s.outputDir, zipFilename)
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	for _, file := range files {
		if err := s.addFileToZip(zipWriter, file); err != nil {
			return err
		}
	}
	return nil
}

func (s *fsStorage) addFileToZip(zipWriter *zip.Writer, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	header.Name = filepath.Base(filename)
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, file)
	return err
}

func (s *fsStorage) DeleteFile(path string) error {
	return os.Remove(path)
}

func (s *fsStorage) DeleteDir(path string) error {
	return os.RemoveAll(path)
}

func (s *fsStorage) ListOutputs() ([]domain.FileInfo, error) {
	files, err := filepath.Glob(filepath.Join(s.outputDir, "*.zip"))
	if err != nil {
		return nil, err
	}

	var results []domain.FileInfo
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}

		results = append(results, domain.FileInfo{
			Name:        filepath.Base(file),
			Size:        info.Size(),
			CreatedAt:   info.ModTime().Format("2006-01-02 15:04:05"),
			DownloadURL: "/download/" + filepath.Base(file),
		})
	}
	return results, nil
}

func (s *fsStorage) GetOutputPath(filename string) string {
	return filepath.Join(s.outputDir, filename)
}

func (s *fsStorage) GetUploadPath(filename string) string {
	return filepath.Join(s.uploadDir, filename)
}
