//go:build embed_aapt

package main

import (
	"os"
	"path/filepath"
	_ "embed"

	"github.com/m33mt33n/apkg"
)

//go:embed aapt.bin
var aapt []byte

func embeded_aapt() (string, uint8) {
	os.MkdirAll(apkg.Data_dir, 0750)
	aapt_path := filepath.Join(apkg.Data_dir, "aapt")
	if err := os.WriteFile(aapt_path, aapt, 0750); err != nil {
		return aapt_path, 64
	}
	return aapt_path, 0
}
