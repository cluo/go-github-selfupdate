package selfupdate

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/ulikunitz/xz"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"
)

func unarchiveTar(src io.Reader, url, cmd string) (io.Reader, error) {
	t := tar.NewReader(src)
	for {
		h, err := t.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("Failed to unarchive .tar file: %s", err)
		}
		_, name := filepath.Split(h.Name)
		if name == cmd {
			return t, nil
		}
	}

	return nil, fmt.Errorf("File '%s' for the command is not found in %s", cmd, url)
}

// UncompressCommand uncompresses the given source. Archive and compression format is
// automatically detected from 'url' parameter, which represents the URL of asset.
// This returns a reader for the uncompressed command given by 'cmd'. '.zip',
// '.tar.gz', '.tar.xz', '.gz' and '.xz' are supported.
func UncompressCommand(src io.Reader, url, cmd string) (io.Reader, error) {
	if strings.HasSuffix(url, ".zip") {
		log.Println("Uncompressing zip file", url)

		// Zip format requires its file size for uncompressing.
		// So we need to read the HTTP response into a buffer at first.
		buf, err := ioutil.ReadAll(src)
		if err != nil {
			return nil, fmt.Errorf("Failed to create buffer for zip file: %s", err)
		}

		r := bytes.NewReader(buf)
		z, err := zip.NewReader(r, r.Size())
		if err != nil {
			return nil, fmt.Errorf("Failed to uncompress zip file: %s", err)
		}

		for _, file := range z.File {
			_, name := filepath.Split(file.Name)
			if !file.FileInfo().IsDir() && name == cmd {
				return file.Open()
			}
		}

		return nil, fmt.Errorf("File '%s' for the command is not found in %s", cmd, url)
	} else if strings.HasSuffix(url, ".tar.gz") {
		log.Println("Uncompressing tar.gz file", url)

		gz, err := gzip.NewReader(src)
		if err != nil {
			return nil, fmt.Errorf("Failed to uncompress .tar.gz file: %s", err)
		}

		return unarchiveTar(gz, url, cmd)
	} else if strings.HasSuffix(url, ".gzip") || strings.HasSuffix(url, ".gz") {
		log.Println("Uncompressing gzip file", url)

		r, err := gzip.NewReader(src)
		if err != nil {
			return nil, fmt.Errorf("Failed to uncompress gzip file downloaded from %s: %s", url, err)
		}

		name := r.Header.Name
		if name != cmd {
			return nil, fmt.Errorf("File name '%s' does not match to command '%s' found in %s", name, cmd, url)
		}

		return r, nil
	} else if strings.HasSuffix(url, ".tar.xz") {
		log.Println("Uncompressing tar.xz file", url)

		xzip, err := xz.NewReader(src)
		if err != nil {
			return nil, fmt.Errorf("Failed to uncompress .tar.xz file: %s", err)
		}

		return unarchiveTar(xzip, url, cmd)
	} else if strings.HasSuffix(url, ".xz") {
		log.Println("Uncompressing xzip file", url)

		xzip, err := xz.NewReader(src)
		if err != nil {
			return nil, fmt.Errorf("Failed to uncompress xzip file downloaded from %s: %s", url, err)
		}
		return xzip, nil
	}

	log.Println("Uncompression is not needed", url)
	return src, nil
}
