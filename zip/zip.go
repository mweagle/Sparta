package zip

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// FileHeaderAnnotator represents a callback function that accepts the current
// file being added to allow it to customize the ZIP archive values
type FileHeaderAnnotator func(header *zip.FileHeader) (*zip.FileHeader, error)

// AnnotateAddToZip is an extended Zip writer that accepts an annotation function
// to customize the FileHeader values written into the archive
func AnnotateAddToZip(zipWriter *zip.Writer,
	source string,
	rootSource string,
	annotator FileHeaderAnnotator,
	logger *logrus.Logger) error {

	linuxZipName := func(platformValue string) string {
		return strings.Replace(platformValue, "\\", "/", -1)
	}

	fullPathSource, err := filepath.Abs(source)
	if nil != err {
		return errors.Wrapf(err, "Failed to get absolute filepath")
	}

	appendFile := func(info os.FileInfo) error {
		zipEntryName := info.Name()
		if rootSource != "" {
			zipEntryName = fmt.Sprintf("%s/%s", linuxZipName(rootSource), info.Name())
		}
		// Create a header for this zipFile, basically let's see
		// if we can get the executable bits to travel along..
		fileHeader, fileHeaderErr := zip.FileInfoHeader(info)
		if fileHeaderErr != nil {
			return fileHeaderErr
		}
		// Update the name to the proper thing...
		fileHeader.Name = zipEntryName
		if annotator != nil {
			annotatedHeader, annotatedHeaderErr := annotator(fileHeader)
			if annotatedHeaderErr != nil {
				return errors.Wrapf(annotatedHeaderErr, "Failed to annotate Zip entry file header")
			}
			fileHeader = annotatedHeader
		}

		// File info for the binary executable
		binaryWriter, binaryWriterErr := zipWriter.CreateHeader(fileHeader)
		if binaryWriterErr != nil {
			return binaryWriterErr
		}
		/* #nosec */
		reader, readerErr := os.Open(fullPathSource)
		if readerErr != nil {
			return readerErr
		}
		written, copyErr := io.Copy(binaryWriter, reader)
		errClose := reader.Close()
		if errClose != nil {
			logger.WithFields(logrus.Fields{
				"Error": errClose,
			}).Warn("Failed to close Zip input stream")
		}
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
			return errors.Wrapf(err, "Failed to create FileInfoHeader")
		}
		// Normalize the Name
		platformName := strings.TrimPrefix(strings.TrimPrefix(path, rootSource), string(os.PathSeparator))
		header.Name = linuxZipName(platformName)

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
		/* #nosec */
		file, err := os.Open(path)
		if err != nil {
			return errors.Wrapf(err, "Failed to open file: %s", path)
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		if err != nil {
			return errors.Wrapf(err, "Failed to copy file contents")
		}
		return nil
	}

	fileInfo, err := os.Stat(fullPathSource)
	if nil != err {
		return errors.Wrapf(err, "Failed to get file information")
	}
	switch mode := fileInfo.Mode(); {
	case mode.IsDir():
		err = filepath.Walk(fullPathSource, directoryWalker)
	case mode.IsRegular():
		err = appendFile(fileInfo)
	default:
		err = errors.New("Inavlid source type")
	}
	if err != nil {
		return errors.Wrapf(err, "Failed to determine file mode")
	}
	return nil
}

// AddToZip creates a source object (either a file, or a directory that will be recursively
// added) to a previously opened zip.Writer.  The archive path of `source` is relative to the
// `rootSource` parameter.
func AddToZip(zipWriter *zip.Writer, source string, rootSource string, logger *logrus.Logger) error {
	return AnnotateAddToZip(zipWriter, source, rootSource, nil, logger)
}
