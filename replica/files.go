package replica

import (
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Files FileInfo
type Files []FileInfo

// We may not neet Sort interface
func (f Files) Len() int      { return len(f) }
func (f Files) Swap(i, j int) { f[i], f[j] = f[j], f[i] }
func (f Files) Less(i, j int) bool {
	if (f[i].IsDir && f[j].IsDir) || (!f[i].IsDir && !f[j].IsDir) {
		return f[i].Name < f[j].Name
	}
	return f[i].IsDir
}

func (f Files) test() string {
	str := ""
	for _, fi := range f {
		str += fi.Name
		if fi.IsDir {
			str += ":dir;"
		} else {
			str += ":file;"
		}
	}
	return str
}

func newFileInfo(resp *http.Response) *FileInfo {
	fi := &FileInfo{metaData: make(map[string]string)}
	fi.Name = filepath.Base(resp.Header.Get("X-Path"))
	fi.Path = resp.Header.Get("X-Path")
	fi.contentType = resp.Header.Get("Content-Type")
	fi.Owner = resp.Header.Get("X-Owner")
	fi.IsDir = resp.Header.Get("X-Type") == "dir"
	if t, err := time.Parse(http.TimeFormat, resp.Header.Get("Last-Modified")); err == nil {
		fi.ModTime = t
	}
	if i, err := strconv.Atoi(resp.Header.Get("X-Replica-Count")); err == nil {
		fi.replicaCount = i
	}
	if i, err := strconv.ParseInt(resp.Header.Get("X-Length"), 10, 64); err == nil {
		fi.Size = i
	}
	for k, v := range resp.Header {
		if !strings.HasPrefix(k, "X-Meta-") {
			continue
		}
		fi.metaData[strings.TrimPrefix(k, "X-Meta-")] = strings.Join(v, " ")
	}
	return fi
}

// FileInfo file information
type FileInfo struct {
	Name         string    `json:"name,omitempty"`
	Path         string    `json:"path,omitempty"`
	Owner        string    `json:"owner,omitempty"`
	IsDir        bool      `json:"is_dir,omitempty"`
	Size         int64     `json:"size,omitempty"`
	ModTime      time.Time `json:"mod_time,omitempty"`
	contentType  string
	replicaCount int
	metaData     map[string]string
}

// ContentType returns contentType of FileInfo
func (f *FileInfo) ContentType() string { return f.contentType }

// ReplicaCount returns replicaCount of FileInfo
func (f *FileInfo) ReplicaCount() int { return f.replicaCount }

// MetaData returns metaData of FileInfo
func (f *FileInfo) MetaData() map[string]string { return f.metaData }
