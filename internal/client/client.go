package client

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"time"
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

type Client struct {
	conn net.Conn
}

func New() (*Client, error) {
	socketPath := getSocketPath()

	conn, err := net.DialTimeout("unix", socketPath, 500*time.Millisecond)
	if err != nil {
		return nil, fmt.Errorf("daemon not running at %s: %w", socketPath, err)
	}

	return &Client{conn: conn}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) GetText() (string, error) {
	resp, err := c.request("get_text", nil)
	if err != nil {
		return "", err
	}
	var text string
	if resp.Data != nil {
		json.Unmarshal(resp.Data, &text)
	}
	return text, nil
}

func (c *Client) SetText(text string) error {
	_, err := c.request("set_text", text)
	return err
}

func (c *Client) GetImage() ([]byte, error) {
	resp, err := c.request("get_image", nil)
	if err != nil {
		return nil, err
	}
	var data []byte
	if resp.Data != nil {
		json.Unmarshal(resp.Data, &data)
	}
	return data, nil
}

func (c *Client) SetImage(pngData []byte) error {
	_, err := c.request("set_image", pngData)
	return err
}

func (c *Client) SetImageType(imageData []byte, mimeType string) error {
	req := struct {
		Data []byte `json:"data"`
		Type string `json:"type"`
	}{
		Data: imageData,
		Type: mimeType,
	}
	_, err := c.request("set_image_type", req)
	return err
}

func (c *Client) GetFiles() ([]string, error) {
	resp, err := c.request("get_files", nil)
	if err != nil {
		return nil, err
	}
	var files []string
	if resp.Data != nil {
		json.Unmarshal(resp.Data, &files)
	}
	return files, nil
}

func (c *Client) SetFiles(files []string) error {
	_, err := c.request("set_files", files)
	return err
}

func (c *Client) request(op string, data interface{}) (*Response, error) {
	id := int(time.Now().UnixNano())

	var dataJSON json.RawMessage
	if data != nil {
		dataJSON, _ = json.Marshal(data)
	}

	req := Request{
		ID:   id,
		Op:   op,
		Data: dataJSON,
	}

	if err := json.NewEncoder(c.conn).Encode(req); err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	var resp Response
	if err := json.NewDecoder(c.conn).Decode(&resp); err != nil {
		return nil, fmt.Errorf("response failed: %w", err)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("%s", *resp.Error)
	}

	return &resp, nil
}

func getSocketPath() string {
	xdgDir := os.Getenv("XDG_RUNTIME_DIR")
	if xdgDir == "" {
		xdgDir = "/tmp"
	}
	return filepath.Join(xdgDir, socketName)
}

func EnsureDaemon() error {
	c, err := New()
	if err == nil {
		c.Close()
		return nil
	}

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot find daemon executable: %w", err)
	}

	daemonPath := filepath.Join(filepath.Dir(exePath), "wlclipd")

	cmd := exec.Command(daemonPath, "-d")
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start daemon: %w", err)
	}

	for i := 0; i < 20; i++ {
		time.Sleep(50 * time.Millisecond)
		c, err := New()
		if err == nil {
			c.Close()
			return nil
		}
	}

	cmd.Process.Kill()
	return fmt.Errorf("daemon failed to start")
}
