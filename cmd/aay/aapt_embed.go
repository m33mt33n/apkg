//go:build embed_aapt

package main

import (
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	_ "embed"

	"github.com/m33mt33n/apkg"
)

//go:embed aapt.bin.gz
var aapt []byte

func embedded_aapt() (string, uint8) {
	os.MkdirAll(apkg.Data_dir, 0750)
	gz_reader, err := gzip.NewReader(
		bytes.NewReader(aapt),
	)
	if err != nil {
		return "", 91
	}
	defer gz_reader.Close()
	aapt_path := filepath.Join(apkg.Data_dir, "aapt")
	var outfile *os.File
	if outfile, err = os.Create(aapt_path); err != nil {
		return aapt_path, 92
	}
	defer outfile.Close()
	if _, err := io.Copy(outfile, gz_reader); err != nil {
		return aapt_path, 93
	}
	outfile.Chmod(0750)
	return aapt_path, 0
}
