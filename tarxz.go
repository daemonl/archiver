package archiver

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ulikunitz/xz"
)

// TarXZ is for TarXZ format
var TarXZ xzFormat

func init() {
	RegisterFormatReaderWriter("TarXZ", TarXZ)
}

type xzFormat struct{}

// Match returns whether filename matches this format.
func (xzFormat) Match(filename string) bool {
	return strings.HasSuffix(strings.ToLower(filename), ".tar.xz") ||
		strings.HasSuffix(strings.ToLower(filename), ".txz") ||
		isTarXz(filename)
}

// isTarXz checks the file has the xz compressed Tar format header by reading
// its beginning block.
func isTarXz(tarxzPath string) bool {
	f, err := os.Open(tarxzPath)
	if err != nil {
		return false
	}
	defer f.Close()

	xzr, err := xz.NewReader(f)
	if err != nil {
		return false
	}

	buf := make([]byte, tarBlockSize)
	n, err := xzr.Read(buf)
	if err != nil || n < tarBlockSize {
		return false
	}

	return hasTarHeader(buf)
}

// MarkWriter the contents of files listed in filePaths in .tar.xz format. File
// paths can be those of regular files or directories.  Regular files are
// stored at the 'root' of the archive, and directories are recursively added.
func (xzFormat) MakeWriter(out io.Writer, filePaths []string, exclusions []string) error {
	xzWriter, err := xz.NewWriter(out)
	if err != nil {
		return fmt.Errorf("error compressing: %v", err)
	}
	defer xzWriter.Close()

	tarWriter := tar.NewWriter(xzWriter)
	defer tarWriter.Close()

	return tarball(filePaths, tarWriter, exclusions)
}

// Open untars source and decompresses the contents into destination.
func (xzFormat) OpenReader(f io.Reader, destination string) error {
	xzReader, err := xz.NewReader(f)
	if err != nil {
		return fmt.Errorf("error decompressing: %v", err)
	}

	return untar(tar.NewReader(xzReader), destination)
}
