package storage

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
)

// zipDir creates a ZIP archive from src contents. Streams without loading
// entire files into RAM. Returns the size of the resulting ZIP file.
func zipDir(src, dstZip string) (int64, error) {
	out, err := os.Create(dstZip)
	if err != nil {
		return 0, err
	}
	defer out.Close()

	zw := zip.NewWriter(out)
	walkErr := filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		// ZIP convention: forward slashes, cross-platform compat.
		zipName := filepath.ToSlash(rel)

		info, err := d.Info()
		if err != nil {
			return err
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = zipName
		header.Method = zip.Deflate

		writer, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(writer, f)
		return err
	})
	if walkErr != nil {
		zw.Close()
		return 0, walkErr
	}
	if err := zw.Close(); err != nil {
		return 0, err
	}

	info, err := os.Stat(dstZip)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}
