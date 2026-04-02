package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"image/png"
	"io"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/jiiihpeeh/wl-clip-go/go/wlclip"
)

const socketName = "wlclipd.sock"

type Request struct {
	ID   int             `json:"id"`
	Op   string          `json:"op"`
	Data json.RawMessage `json:"data,omitempty"`
}

type Response struct {
	ID    int             `json:"id"`
	Data  json.RawMessage `json:"data,omitempty"`
	Error *string         `json:"error,omitempty"`
}

func main() {
	daemonMode := flag.Bool("d", false, "Run as daemon")
	flag.Parse()

	if !*daemonMode {
		fmt.Fprintf(os.Stderr, "wlclipd: run with -d flag to start daemon\n")
		os.Exit(1)
	}

	socketPath := getSocketPath()

	os.Remove(socketPath)

	ln, err := net.Listen("unix", socketPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "wlclipd: failed to listen: %v\n", err)
		os.Exit(1)
	}
	defer os.Remove(socketPath)

	if err := os.Chmod(socketPath, 0600); err != nil {
		fmt.Fprintf(os.Stderr, "wlclipd: failed to set socket permissions: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "wlclipd: listening on %s\n", socketPath)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		os.Remove(socketPath)
		os.Exit(0)
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	dec := json.NewDecoder(conn)
	enc := json.NewEncoder(conn)

	for {
		var req Request
		if err := dec.Decode(&req); err != nil {
			if err != io.EOF {
				fmt.Fprintf(os.Stderr, "wlclipd: decode error: %v\n", err)
			}
			return
		}

		resp := handleRequest(req)
		if err := enc.Encode(resp); err != nil {
			fmt.Fprintf(os.Stderr, "wlclipd: encode error: %v\n", err)
			return
		}
	}
}

func handleRequest(req Request) Response {
	var resp Response
	resp.ID = req.ID

	switch req.Op {
	case "get_text":
		text, err := wlclip.GetText()
		if err != nil {
			errStr := err.Error()
			resp.Error = &errStr
		} else {
			data, _ := json.Marshal(text)
			resp.Data = data
		}

	case "set_text":
		var text string
		if req.Data != nil {
			json.Unmarshal(req.Data, &text)
		}
		if err := wlclip.SetText(text); err != nil {
			errStr := err.Error()
			resp.Error = &errStr
		}

	case "get_image":
		img, err := wlclip.GetImage()
		if err != nil {
			errStr := err.Error()
			resp.Error = &errStr
		} else {
			var buf bytes.Buffer
			if err := png.Encode(&buf, img); err != nil {
				errStr := err.Error()
				resp.Error = &errStr
			} else {
				data, _ := json.Marshal(buf.Bytes())
				resp.Data = data
			}
		}

	case "set_image":
		var pngData []byte
		if req.Data != nil {
			json.Unmarshal(req.Data, &pngData)
		}
		if err := wlclip.SetImageType(pngData, "image/png"); err != nil {
			errStr := err.Error()
			resp.Error = &errStr
		}

	case "set_image_type":
		var reqData struct {
			Data string `json:"data"`
			Type string `json:"type"`
		}
		if req.Data != nil {
			json.Unmarshal(req.Data, &reqData)
		}
		imageData := []byte(reqData.Data)
		if err := wlclip.SetImageType(imageData, reqData.Type); err != nil {
			errStr := err.Error()
			resp.Error = &errStr
		}

	case "get_files":
		files, err := wlclip.GetFiles()
		if err != nil {
			errStr := err.Error()
			resp.Error = &errStr
		} else {
			data, _ := json.Marshal(files)
			resp.Data = data
		}

	case "set_files":
		var files []string
		if req.Data != nil {
			json.Unmarshal(req.Data, &files)
		}
		if err := wlclip.SetFiles(files); err != nil {
			errStr := err.Error()
			resp.Error = &errStr
		}

	default:
		errStr := fmt.Sprintf("unknown operation: %s", req.Op)
		resp.Error = &errStr
	}

	return resp
}

func getSocketPath() string {
	xdgDir := os.Getenv("XDG_RUNTIME_DIR")
	if xdgDir == "" {
		xdgDir = "/tmp"
	}
	return filepath.Join(xdgDir, socketName)
}
