package archiver

import (
	"archive/tar"
	"io"
	"os"
	"strings"

	"github.com/golang/snappy"
)

// TarSz is for TarSz format
var TarSz tarSzFormat

func init() {
	RegisterFormatReaderWriter("TarSz", TarSz)
}

type tarSzFormat struct{}

func (tarSzFormat) Match(filename string) bool {
	return strings.HasSuffix(strings.ToLower(filename), ".tar.sz") || strings.HasSuffix(strings.ToLower(filename), ".tsz") || isTarSz(filename)
}

// isTarSz checks the file has the sz compressed Tar format header by
// reading its beginning block.
func isTarSz(tarszPath string) bool {
	f, err := os.Open(tarszPath)
	if err != nil {
		return false
	}
	defer f.Close()

	szr := snappy.NewReader(f)
	buf := make([]byte, tarBlockSize)
	n, err := szr.Read(buf)
	if err != nil || n < tarBlockSize {
		return false
	}

	return hasTarHeader(buf)
}

// MakeWriter writes the contents of files listed in filePaths in .tar.sz
// format. File paths can be those of regular files or directories. Regular
// files are stored at the 'root' of the archive, and directories are
// recursively added.
func (tarSzFormat) MakeWriter(out io.Writer, filePaths []string, exclusions []string) error {
	szWriter := snappy.NewBufferedWriter(out)
	defer szWriter.Close()

	tarWriter := tar.NewWriter(szWriter)
	defer tarWriter.Close()

	return tarball(filePaths, tarWriter, exclusions)
}

// OpenReader untars source and decompresses the contents into destination.
func (tarSzFormat) OpenReader(f io.Reader, destination string) error {
	szr := snappy.NewReader(f)
	return untar(tar.NewReader(szr), destination)
}
