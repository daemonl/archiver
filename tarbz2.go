package archiver

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/dsnet/compress/bzip2"
)

// TarBz2 is for TarBz2 format
var TarBz2 tarBz2Format

func init() {
	RegisterFormatReaderWriter("TarBz2", TarBz2)
}

type tarBz2Format struct{}

func (tarBz2Format) Match(filename string) bool {
	return strings.HasSuffix(strings.ToLower(filename), ".tar.bz2") ||
		strings.HasSuffix(strings.ToLower(filename), ".tbz2") ||
		isTarBz2(filename)
}

// isTarBz2 checks the file has the bzip2 compressed Tar format header by
// reading its beginning block.
func isTarBz2(tarbz2Path string) bool {
	f, err := os.Open(tarbz2Path)
	if err != nil {
		return false
	}
	defer f.Close()

	bz2r, err := bzip2.NewReader(f, nil)
	if err != nil {
		return false
	}
	defer bz2r.Close()

	buf := make([]byte, tarBlockSize)
	n, err := bz2r.Read(buf)
	if err != nil || n < tarBlockSize {
		return false
	}

	return hasTarHeader(buf)
}

// MakeWriter writes the contents of files listed in filePaths to the writer in
// tar.bz2 format. File paths can be those of regular files or directories.
// Regular files are stored at the 'root' of the archive, and directories are
// recursively added.
func (tarBz2Format) MakeWriter(out io.Writer, filePaths []string, exclusions []string) error {
	bz2Writer, err := bzip2.NewWriter(out, nil)
	if err != nil {
		return fmt.Errorf("error compressing: %v", err)
	}
	defer bz2Writer.Close()

	tarWriter := tar.NewWriter(bz2Writer)
	defer tarWriter.Close()

	return tarball(filePaths, tarWriter, exclusions)
}

// OpenReader untars source and decompresses the contents into destination.
func (tarBz2Format) OpenReader(f io.Reader, destination string) error {
	bz2r, err := bzip2.NewReader(f, nil)
	if err != nil {
		return fmt.Errorf("error decompressing: %v", err)
	}
	defer bz2r.Close()

	return untar(tar.NewReader(bz2r), destination)
}
