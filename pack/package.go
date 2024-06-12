package pack

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gabriel-vasile/mimetype"
)

func Install(packageUrl, destPath string) error {
	tmpFile, err := Download(packageUrl)
	if err != nil {
		return fmt.Errorf("failed downloading %s: %w", packageUrl, err)
	}
	defer os.Remove(tmpFile)

	err = UnpackFile(tmpFile, destPath)
	if err != nil {
		return fmt.Errorf("failed unpacking %q: %w", tmpFile, err)
	}
	return nil
}

func Download(packageUrl string) (string, error) {
	res, err := http.Get(packageUrl)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	contentType := res.Header.Get("Content-Type")
	if contentType == "" {
		return "", errors.New("content type header not found")
	}

	mime := mimetype.Lookup(contentType)
	if mime == nil {
		return "", fmt.Errorf("no extension found to file of type %q", contentType)
	}

	fileNameFmt := "download-*"
	if ext := mime.Extension(); ext != "" {
		fileNameFmt += ext
	}

	tmpFile, err := os.CreateTemp(".", fileNameFmt)
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	tmpFile.ReadFrom(res.Body)
	return tmpFile.Name(), nil
}

func UnpackFile(archivePath, outputDir string) error {
	mime, err := mimetype.DetectFile(archivePath)
	if err != nil {
		return err
	}

	archiveFile, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer archiveFile.Close()

	if err := unpack(outputDir, mime, archiveFile); err != nil {
		return err
	}
	return nil
}

func isArchive(mime *mimetype.MIME) bool {
	return mime.Is("application/gzip") || mime.Is("application/x-tar")
}

func unpack(output string, mime *mimetype.MIME, reader io.Reader) error {
	if mime.Is("application/gzip") {
		return UnpackGzip(output, reader)
	}

	if mime.Is("application/x-tar") {
		return UnpackTar(output, reader)
	}

	return errors.New("unsupported archive type %q")
}

func UnpackTar(outputDir string, reader io.Reader) error {
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return err
	}

	tarReader := tar.NewReader(reader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		dstFilePath := filepath.Join(outputDir, header.Name)

		if header.FileInfo().IsDir() {
			if err := os.MkdirAll(dstFilePath, header.FileInfo().Mode()); err != nil {
				return err
			}
			continue
		}

		if parentFolder := filepath.Dir(dstFilePath); parentFolder != "" {
			if err := os.MkdirAll(parentFolder, os.ModePerm); err != nil {
				return err
			}
		}

		newFile, err := os.Create(dstFilePath)
		if err != nil {
			return err
		}
		defer newFile.Close()

		_, err = io.Copy(newFile, tarReader)
		if err != nil {
			return err
		}
		os.Chmod(dstFilePath, header.FileInfo().Mode())
	}
	return removeIntermediateDir(outputDir)
}

func removeIntermediateDir(rootPath string) error {
	innerFolder := rootPath
	var walkedFolders []string

	for {
		folderEntries, err := os.ReadDir(innerFolder)
		if err != nil {
			return err
		}

		if len(folderEntries) == 0 {
			break
		}

		if len(folderEntries) == 1 && folderEntries[0].IsDir() {
			walkedFolders = append(walkedFolders, folderEntries[0].Name())
			innerFolder = filepath.Join(innerFolder, folderEntries[0].Name())
			continue
		}

		for _, entry := range folderEntries {
			entryPath := filepath.Join(innerFolder, entry.Name())
			newPath := filepath.Join(rootPath, entry.Name())

			if err := os.Rename(entryPath, newPath); err != nil {
				return err
			}
		}

		removePath := rootPath
		for _, oldFolder := range walkedFolders {
			removePath = filepath.Join(removePath, oldFolder)
			if os.Remove(removePath) != nil {
				break
			}
		}
		break
	}
	return nil
}

type bufferedReader struct {
	reader      io.Reader
	buffer      []byte
	pos         int
	isBuffering bool
}

func newBufferedReader(reader io.Reader) bufferedReader {
	return bufferedReader{
		reader,
		make([]byte, 0, 256),
		0,
		true,
	}
}

func (buffReader *bufferedReader) Read(data []byte) (int, error) {
	if buffReader.isBuffering {
		n, err := buffReader.reader.Read(data)
		if n > 0 && (err == nil || err == io.EOF) {
			buffReader.buffer = append(buffReader.buffer, data[:n]...)
		}
		return n, err
	}

	readBytes := 0
	index := 0
	for ; index < len(data); index++ {
		if index+buffReader.pos >= len(buffReader.buffer) {
			break
		}
		data[index] = buffReader.buffer[index+buffReader.pos]
		readBytes++
	}
	buffReader.pos += index

	if readBytes < len(data) {
		n, err := buffReader.reader.Read(data[readBytes:])
		readBytes += n

		return readBytes, err
	}
	return readBytes, nil
}

func (buffReader *bufferedReader) SetBuffMode(isBuffering bool) {
	buffReader.isBuffering = isBuffering
}

func UnpackGzip(outputPath string, reader io.Reader) error {
	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	// use a buffered reader store all bytes read by the MIME detector
	buffReader := newBufferedReader(gzipReader)
	mime, err := mimetype.DetectReader(&buffReader)
	if err != nil {
		return err
	}
	// change reader mode to consume stored bytes
	buffReader.SetBuffMode(false)

	if isArchive(mime) {
		return unpack(outputPath, mime, &buffReader)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, &buffReader)
	return err
}
