package archiver

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strings"
)

// TarGz is for TarGz format
var TarGz tarGzFormat

func init() {
	RegisterFormatReaderWriter("TarGz", TarGz)
}

type tarGzFormat struct{}

func (tarGzFormat) Match(filename string) bool {
	return strings.HasSuffix(strings.ToLower(filename), ".tar.gz") ||
		strings.HasSuffix(strings.ToLower(filename), ".tgz") ||
		isTarGz(filename)
}

// isTarGz checks the file has the gzip compressed Tar format header by reading
// its beginning block.
func isTarGz(targzPath string) bool {
	f, err := os.Open(targzPath)
	if err != nil {
		return false
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return false
	}
	defer gzr.Close()

	buf := make([]byte, tarBlockSize)
	n, err := gzr.Read(buf)
	if err != nil || n < tarBlockSize {
		return false
	}

	return hasTarHeader(buf)
}

// MakeWriter writes the contents of files listed in filePaths to the writer in
// .tar.gz format. It works the same way Tar does, but with gzip compression.
func (tarGzFormat) MakeWriter(out io.Writer, filePaths []string, exclusions []string) error {
	gzWriter := gzip.NewWriter(out)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	return tarball(filePaths, tarWriter, exclusions)
}

// OpenReader untars source and decompresses the contents into destination.
func (tarGzFormat) OpenReader(f io.Reader, destination string) error {
	gzr, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("create new gzip reader: %v", err)
	}
	defer gzr.Close()

	return untar(tar.NewReader(gzr), destination)
}
