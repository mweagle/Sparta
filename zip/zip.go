package zip

import (
	"archive/zip"
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// AddToZip creates a source object (either a file, or a directory that will be recursively
// added) to a previously opened zip.Writer.  The archive path of `source` is relative to the
// `rootSource` parameter.
func AddToZip(zipWriter *zip.Writer, source string, rootSource string, logger *logrus.Logger) error {
	fullPathSource, err := filepath.Abs(source)
	if nil != err {
		return err
	}
	appendFile := func(sourceFile string) error {
		// Get the relative path
		var name = filepath.Base(sourceFile)
		if sourceFile != rootSource {
			name = strings.TrimPrefix(strings.TrimPrefix(sourceFile, rootSource), string(os.PathSeparator))

			// Normalize the name s.t. path delimiters are AWS Linux friendly when
			// unpacking the archive.
			name = strings.Replace(name, "\\", "/", -1)
		}
		binaryWriter, errCreate := zipWriter.Create(name)
		if errCreate != nil {
			return fmt.Errorf("Failed to create ZIP entry: %s", filepath.Base(sourceFile))
		}
		reader, errOpen := os.Open(sourceFile)
		if errOpen != nil {
			return fmt.Errorf("Failed to open file: %s", sourceFile)
		}
		defer reader.Close()
		io.Copy(binaryWriter, reader)
		logger.WithFields(logrus.Fields{
			"Path": sourceFile,
		}).Debug("Archiving file")

		return nil
	}

	directoryWalker := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = strings.TrimPrefix(strings.TrimPrefix(path, rootSource), string(os.PathSeparator))
		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	}

	fileInfo, err := os.Stat(fullPathSource)
	if nil != err {
		return err
	}
	switch mode := fileInfo.Mode(); {
	case mode.IsDir():
		err = filepath.Walk(fullPathSource, directoryWalker)
	case mode.IsRegular():
		err = appendFile(fullPathSource)
	default:
		err = errors.New("Inavlid source type")
	}
	zipWriter.Close()
	return err
}
