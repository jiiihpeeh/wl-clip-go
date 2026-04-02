//go:build ignore

package main

import (
	"fmt"
	"os"
	"strings"
)

type ArchConfig struct {
	GOARCH    string
	BuildTags string
	LibPath   string
}

func main() {
	data, err := os.ReadFile("clip_src.go")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read clip_src.go: %v\n", err)
		os.Exit(1)
	}
	src := string(data)

	archs := []ArchConfig{
		{
			GOARCH:    "amd64",
			BuildTags: "//go:build (amd64 || 386) && !arm64\n// +build amd64,!arm64",
			LibPath:   "x86_64-unknown-linux-gnu",
		},
		{
			GOARCH:    "arm64",
			BuildTags: "//go:build arm64\n// +build arm64",
			LibPath:   "aarch64-unknown-linux-gnu",
		},
	}

	for _, arch := range archs {
		content := src
		buildTags := buildTagsFor(arch)
		archLibPath := arch.LibPath

		lines := strings.Split(content, "\n")
		var outLines []string
		skipMode := true
		for _, line := range lines {
			if skipMode && (strings.TrimSpace(line) == "//go:build ignore" || line == "") {
				if strings.TrimSpace(line) == "//go:build ignore" {
					outLines = append(outLines, buildTags)
					outLines = append(outLines, "") // blank line after build tags
				}
				continue
			}
			skipMode = false
			line = strings.ReplaceAll(line, "ARCH_LIB_PATH", archLibPath)
			outLines = append(outLines, line)
		}
		content = strings.Join(outLines, "\n")

		filename := fmt.Sprintf("clip_%s.go", arch.GOARCH)
		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write %s: %v\n", filename, err)
			os.Exit(1)
		}
		fmt.Printf("Generated %s\n", filename)
	}
}

func buildTagsFor(arch ArchConfig) string {
	if arch.GOARCH == "amd64" {
		return "//go:build (amd64 || 386) && !arm64"
	}
	return "//go:build arm64"
}
