package replica

import (
	//"fmt"
	"sort"
	"testing"
)

func TestFiles(t *testing.T) {
	files := []FileInfo{
		FileInfo{Name: "B", IsDir: false},
		FileInfo{Name: "A", IsDir: true},
		FileInfo{Name: "A", IsDir: false},
		FileInfo{Name: "B", IsDir: true},
	}
	fls := Files(files)
	sort.Sort(fls)
	if fls.test() != "A:dir;B:dir;A:file;B:file;" {
		t.Errorf("sort failed, got %s", fls.test())
	}
}

func TestMetaData(t *testing.T) {
	fi := &FileInfo{}
	if fi.MetaData()["test"] != "" {
		t.Error("expected empty value got ", fi.MetaData()["test"])
	}
}
