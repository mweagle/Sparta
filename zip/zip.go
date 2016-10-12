package zip

import (
	"archive/zip"
	"errors"
	"github.com/Sirupsen/logrus"
	"io"
	"os"
	"path"
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

	appendFile := func(info os.FileInfo) error {
		zipEntryName := source
		if "" != rootSource {
			zipEntryName = path.Join(strings.Replace(rootSource, "\\", "/", -1), info.Name())
		}
		// File info for the binary executable
		binaryWriter, binaryWriterErr := zipWriter.Create(zipEntryName)
		if binaryWriterErr != nil {
			return binaryWriterErr
		}
		reader, readerErr := os.Open(fullPathSource)
		if readerErr != nil {
			return readerErr
		}
		written, copyErr := io.Copy(binaryWriter, reader)
		reader.Close()

		logger.WithFields(logrus.Fields{
			"WrittenBytes": written,
			"SourcePath":   fullPathSource,
			"ZipName":      zipEntryName,
		}).Debug("Archiving file")
		return copyErr
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
		err = appendFile(fileInfo)
	default:
		err = errors.New("Inavlid source type")
	}
	return err
}
