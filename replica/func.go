package replica

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const sniffLen = 512

// GetInfo makes HEAD request to get FileInfo of resource
func (c *Client) GetInfo(name string) (*FileInfo, error) {
	req, err := c.newRequest("HEAD", name, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	fi := newFileInfo(resp)
	return fi, nil
}

// Get makes GET request to get a resource
// if resource is directory than Files object returned
// if resource is file than bytes array returned
func (c *Client) Get(name string) (io.ReadCloser, Files, error) {
	req, err := c.newRequest("GET", name, nil)
	if err != nil {
		return nil, nil, err
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, nil, err
	}
	if resp.Header.Get("X-Type") == "dir" {
		fls := Files{}
		err = json.NewDecoder(resp.Body).Decode(&fls)
		return nil, fls, err
	}

	return resp.Body, nil, nil
}

// CreateFile makes PUT request to create a resource
func (c *Client) CreateFile(name string, fi *FileInfo, read io.Reader) (err error) {
	req, err := c.newRequest("PUT", name, read)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", fi.ContentType())
	req.ContentLength = fi.Size
	if fi.replicaCount > 0 {
		req.Header.Add("X-Replica-Count", fmt.Sprint(fi.ReplicaCount()))
	}
	for k, v := range fi.metaData {
		req.Header.Add("X-Meta-"+strings.Title(k), v)
	}
	_, err = c.do(req)
	return err
}

// CreateDir makes PUT request to create a directory
func (c *Client) CreateDir(name string, rep int, meta map[string]string) (err error) {
	fi := &FileInfo{
		replicaCount: rep,
		contentType:  "application/x-directory",
		metaData:     meta,
	}
	return c.CreateFile(name, fi, nil)
}

// Remove makes DELETE request to delete a resource
func (c *Client) Remove(name string) (err error) {
	req, err := c.newRequest("DELETE", name, nil)
	if err != nil {
		return
	}
	_, err = c.do(req)
	return err
}

// RemoveAll makes DELETE request to delete a resource recursivly
func (c *Client) RemoveAll(name string) (err error) {
	req, err := c.newRequest("DELETE", name, nil)
	if err != nil {
		return
	}
	req.Header.Add("X-Remove-All", "x")
	_, err = c.do(req)
	return err
}

// Exist makes OPTIONS request to check resource existence
func (c *Client) Exist(name string) (err error) {
	req, err := c.newRequest("OPTIONS", name, nil)
	if err != nil {
		return
	}
	_, err = c.do(req)
	return err
}

// Update makes POST request to change resources metadata
func (c *Client) Update(name string, meta, rmeta map[string]string) error {
	req, err := c.newRequest("POST", name, nil)
	if err != nil {
		return err
	}
	for k, v := range meta {
		req.Header.Add("X-Meta-"+strings.Title(k), v)
	}
	for k, v := range rmeta {
		req.Header.Add("X-Remove-Meta-"+strings.Title(k), v)
	}
	_, err = c.do(req)
	return err
}

// OpenFile opens a file to read and returns files info
func OpenFile(name string, meta map[string]string) (*FileInfo, io.ReadCloser, error) {
	fi := &FileInfo{metaData: make(map[string]string)}
	f, err := os.Stat(name)
	if err != nil {
		return nil, nil, err
	}
	fi.Size = f.Size()
	ctype := mime.TypeByExtension(filepath.Ext(name))
	rdc, err := os.Open(name)
	if err != nil {
		return nil, nil, err
	}
	if ctype == "" {
		// read a chunk to decide between utf-8 text and binary
		var buf [sniffLen]byte
		n, _ := io.ReadFull(rdc, buf[:])
		ctype = http.DetectContentType(buf[:n])
		_, err = rdc.Seek(0, os.SEEK_SET)
	}
	fi.contentType = ctype
	fi.metaData = meta
	return fi, rdc, nil
}
