package archiver

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

// Archiver represent a archive format
type Archiver interface {
	// Match checks supported files
	Match(filename string) bool
	// Make makes an archive.
	Make(destination string, sources []string) error
	// MakeWriter writes an archive to an io.Write
	MakeWriter(destination io.Writer, sources []string, exclude []string) error
	// Open extracts an archive.
	Open(source, destination string) error
	// OpenReader extracts an archive from an io.Reader
	OpenReader(source io.Reader, destination string) error
}

type ArchiverReaderWriter interface {
	// Match checks supported files
	Match(filename string) bool
	// MakeWriter writes an archive to an io.Write
	MakeWriter(destination io.Writer, sources []string, exclude []string) error
	// OpenReader extracts an archive from an io.Reader
	OpenReader(source io.Reader, destination string) error
}

type archiverReaderWriterExtend struct {
	ArchiverReaderWriter
}

func (format archiverReaderWriterExtend) Make(destination string, sources []string) error {
	out, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("error creating %s: %v", destination, err)
	}
	defer out.Close()
	return format.MakeWriter(out, sources, []string{destination})
}

func (format archiverReaderWriterExtend) Open(source, destination string) error {

	f, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("%s: failed to open archive: %v", source, err)
	}
	defer f.Close()
	return format.OpenReader(f, destination)
}

// SupportedFormats contains all supported archive formats
var SupportedFormats = map[string]Archiver{}

// RegisterFormat adds a supported archive format
func RegisterFormat(name string, format Archiver) {
	if _, ok := SupportedFormats[name]; ok {
		log.Printf("Format %s already exists, skip!\n", name)
		return
	}
	SupportedFormats[name] = format
}

func RegisterFormatReaderWriter(name string, format ArchiverReaderWriter) {
	RegisterFormat(name, archiverReaderWriterExtend{
		ArchiverReaderWriter: format,
	})
}

func writeNewFile(fpath string, in io.Reader, fm os.FileMode) error {
	err := os.MkdirAll(filepath.Dir(fpath), 0755)
	if err != nil {
		return fmt.Errorf("%s: making directory for file: %v", fpath, err)
	}

	out, err := os.Create(fpath)
	if err != nil {
		return fmt.Errorf("%s: creating new file: %v", fpath, err)
	}
	defer out.Close()

	err = out.Chmod(fm)
	if err != nil && runtime.GOOS != "windows" {
		return fmt.Errorf("%s: changing file mode: %v", fpath, err)
	}

	_, err = io.Copy(out, in)
	if err != nil {
		return fmt.Errorf("%s: writing file: %v", fpath, err)
	}
	return nil
}

func writeNewSymbolicLink(fpath string, target string) error {
	err := os.MkdirAll(filepath.Dir(fpath), 0755)
	if err != nil {
		return fmt.Errorf("%s: making directory for file: %v", fpath, err)
	}

	err = os.Symlink(target, fpath)
	if err != nil {
		return fmt.Errorf("%s: making symbolic link for: %v", fpath, err)
	}

	return nil
}

func writeNewHardLink(fpath string, target string) error {
	err := os.MkdirAll(filepath.Dir(fpath), 0755)
	if err != nil {
		return fmt.Errorf("%s: making directory for file: %v", fpath, err)
	}

	err = os.Link(target, fpath)
	if err != nil {
		return fmt.Errorf("%s: making hard link for: %v", fpath, err)
	}

	return nil
}

func mkdir(dirPath string) error {
	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		return fmt.Errorf("%s: making directory: %v", dirPath, err)
	}
	return nil
}
