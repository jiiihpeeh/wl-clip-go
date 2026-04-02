package main

import (
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
)

func main() {
	libsDir := "rust/libs"

	archs := []string{"x86_64-unknown-linux-gnu", "aarch64-unknown-linux-gnu"}

	for _, arch := range archs {
		archDir := filepath.Join(libsDir, arch)
		aPath := filepath.Join(archDir, "libwlclip.a")
		gzPath := filepath.Join(archDir, "libwlclip.a.gz")

		if _, err := os.Stat(aPath); os.IsNotExist(err) {
			if _, err := os.Stat(gzPath); err == nil {
				if err := decompress(gzPath, aPath); err != nil {
					panic(err)
				}
			}
		}
	}
}

func decompress(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	r, err := gzip.NewReader(in)
	if err != nil {
		return err
	}
	defer r.Close()

	_, err = io.Copy(out, r)
	return err
}
