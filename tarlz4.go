package archiver

import (
	"archive/tar"
	"io"
	"os"
	"strings"

	"github.com/pierrec/lz4"
)

// TarLz4 is for TarLz4 format
var TarLz4 tarLz4Format

func init() {
	RegisterFormatReaderWriter("TarLz4", TarLz4)
}

type tarLz4Format struct{}

func (tarLz4Format) Match(filename string) bool {
	return strings.HasSuffix(strings.ToLower(filename), ".tar.lz4") || strings.HasSuffix(strings.ToLower(filename), ".tlz4") || isTarLz4(filename)
}

// isTarLz4 checks the file has the lz4 compressed Tar format header by
// reading its beginning block.
func isTarLz4(tarlz4Path string) bool {
	f, err := os.Open(tarlz4Path)
	if err != nil {
		return false
	}
	defer f.Close()

	lz4r := lz4.NewReader(f)
	buf := make([]byte, tarBlockSize)
	n, err := lz4r.Read(buf)
	if err != nil || n < tarBlockSize {
		return false
	}

	return hasTarHeader(buf)
}

// MakeWriter writes the contents of files listed in filePaths to the writer in
// .tar.lz4 format. File paths can be those of regular files or directories.
// Regular files are stored at the 'root' of the archive, and directories are
// recursively added.
func (tarLz4Format) MakeWriter(out io.Writer, filePaths []string, exclusions []string) error {
	lz4Writer := lz4.NewWriter(out)
	defer lz4Writer.Close()

	tarWriter := tar.NewWriter(lz4Writer)
	defer tarWriter.Close()

	return tarball(filePaths, tarWriter, exclusions)
}

// OpenReader untars source and decompresses the contents into destination.
func (tarLz4Format) OpenReader(f io.Reader, destination string) error {
	lz4r := lz4.NewReader(f)
	return untar(tar.NewReader(lz4r), destination)
}
