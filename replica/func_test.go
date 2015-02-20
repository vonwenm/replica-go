package replica

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"sort"
	"testing"
	"time"
)

const (
	create = 1
	update = 2
	remove = 4
)

type tmeta struct {
	name  string
	value string
	state int
}

var testfiles = []struct {
	name    string
	md5sum  string
	path    string
	contype string
	rc      int
	size    int64
	meta    []tmeta
}{
	{
		name:    "gopherblue.png",
		md5sum:  "ca1f746d6f232f87fca4e4d94ef6f3ab",
		path:    "public/two/gopherblue.png",
		contype: "image/png",
		rc:      1,
		size:    70372,
		meta: []tmeta{
			{"Color", "blue", update},
			{"Animal", "dog", create},
			{"Size", "big", remove},
		},
	},
	{
		name:    "gophers.png",
		md5sum:  "d93d15898447ba7e1504dd6205c6c47b",
		path:    "public/two/gophers.png",
		contype: "image/png",
		rc:      2,
		size:    8042,
		meta: []tmeta{
			{"Color", "n/a", create},
			{"Animal", "cat", remove},
			{"Size", "small", update},
		},
	},
	{
		name:    "gophercolor.png",
		md5sum:  "3091ffa4bfa94e3b1f415e0f8980cb14",
		path:    "one/gophercolor.png",
		contype: "image/png",
		rc:      2,
		size:    54152,
		meta: []tmeta{
			{"Color", "multi", remove},
			{"Animal", "bird", update},
			{"Size", "mid", create},
		},
	},
	{
		name:    "test",
		md5sum:  "22db44c7d15ca94e90c112008c97e5e4",
		path:    "one/test",
		contype: "application/octet-stream",
		rc:      1,
		size:    819200,
		meta: []tmeta{
			{"Color", "none", remove},
			{"Animal", "bird", create},
			{"Size", "biggest", update},
		},
	},
}

func metaGen(t []tmeta, tp int) map[string]string {
	r := make(map[string]string)
	for _, m := range t {
		if m.state&tp > 0 {
			r[m.name] = m.value
		}
	}
	return r
}

func TestToken(t *testing.T) {
	initServer()
	initClient()

	tk, err := client.Token()
	if err != nil {
		t.Fatal(err)
	}
	if tk.String() != token.String() {
		t.Errorf("token expected %s, got %s", token, tk)
	}
	if tk.Expires != token.Expires {
		t.Errorf("token expire expected %d, got %d", token.Expires, tk.Expires)
	}
	// get current token
	tkc, err := client.Token()
	if err != nil {
		t.Error(err)
	}
	if (tk != tkc) || (tkc != client.token) {
		t.Errorf("failed tk %p tkc %p client.token %p", tk, tkc, client.token)
	}
}

func TestCreateDir(t *testing.T) {
	err := client.CreateDir("test/will/fail", 2, nil)
	if err == nil {
		t.Fatal("expected not found error")
	}

	err = client.CreateDir("public/two", 2, nil)
	if err != nil {
		t.Fatal(err)
	}

	err = client.CreateDir("one", 1, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateFile(t *testing.T) {
	err := client.CreateFile("path", &FileInfo{contentType: "application/json"}, nil)
	if err == nil {
		t.Error("expected length required error, got <nil>")
	}
	if err = client.CreateFile("%", nil, nil); err == nil {
		t.Error("expected error, got <nil>")
	}
	for _, f := range testfiles {
		fi, file, err := OpenFile("test_files/"+f.name, metaGen(f.meta, create|remove))
		if err != nil {
			t.Fatal(err)
		}
		fi.replicaCount = f.rc

		err = client.CreateFile(f.path, fi, file)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestListDir(t *testing.T) {
	_, files, err := client.Get("public/two")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Errorf("files count expected 2, got %d", len(files))
	}
	sort.Sort(files)
	_, files2, err := client.Get("one")
	if err != nil {
		t.Error(err)
	}

	if len(files2) != 2 {
		t.Errorf("files count expected 2, got %d", len(files))
	}
	sort.Sort(files2)

	files = append(files, files2...)
	for i, f := range testfiles {
		fi := files[i]
		if fi.IsDir {
			t.Error("expected type dir, got directory")
		}
		now := time.Now().Format(time.RFC822)
		mt := fi.ModTime.Format(time.RFC822)
		if mt != now {
			t.Errorf("expected ModTime %s, got %s", now, mt)
		}
		if fi.Name != f.name {
			t.Errorf("expected name %s, got %s", f.name, fi.Name)
		}
		if fi.Owner != "test" {
			t.Errorf("expected owner test, got %s", fi.Owner)
		}
		if fi.Path != f.path {
			t.Errorf("expected path %s, got %s", f.path, fi.Path)
		}
		if fi.Size != f.size {
			t.Errorf("expected size %d, got %d", f.size, fi.Size)
		}
	}
}

func TestGetFile(t *testing.T) {
	for _, f := range testfiles {
		rcls, _, err := client.Get(f.path)
		if err != nil {
			t.Fatal(err)
		}

		m := md5.New()
		_, err = io.Copy(m, rcls)
		if err != nil {
			t.Fatal(err)
		}

		s := fmt.Sprintf("%x", m.Sum(nil))
		if s != f.md5sum {
			t.Errorf("Sum do not match expected %s got %s", f.md5sum, s)
		}
	}
}

func TestGetFail(t *testing.T) {
	if _, _, err := client.Get("%"); err == nil {
		t.Error("expected no error, got ", err)
	}
	if _, _, err := client.Get("notfound"); err == nil {
		t.Error("expected no error, got ", err)
	}
}

func TestExist(t *testing.T) {
	if err := client.Exist("one"); err != nil {
		t.Error(err)
	}
	if err := client.Exist(testfiles[0].path); err != nil {
		t.Error(err)
	}
	if err := client.Exist("not/exist"); err == nil {
		t.Error("expected err, got nil")
	}
	if err := client.Exist("%"); err == nil {
		t.Error("expected err, got nil")
	}
}

func TestUpdate(t *testing.T) {
	for _, f := range testfiles {
		err := client.Update(f.path, metaGen(f.meta, update), metaGen(f.meta, remove))
		if err != nil {
			t.Error(err)
		}
	}
	// fail
	if err := client.Update("%", nil, nil); err == nil {
		t.Error("expected error got <nil>")
	}
}

func TestGetInfo(t *testing.T) {
	for _, f := range testfiles {
		fi, err := client.GetInfo(f.path)
		if err != nil {
			t.Error(err)
			continue
		}
		if fi.ContentType() != f.contype {
			t.Errorf("expected Content-Type %s, got %s", f.contype, fi.ContentType())
		}
		if fi.IsDir {
			t.Error("expected type dir, got directory")
		}
		now := time.Now().Format("2006 01 02T15:04")
		mt := fi.ModTime.Format("2006 01 02T15:04")
		if mt != now {
			t.Errorf("expected ModTime %s, got %s", now, mt)
		}
		if fi.Name != f.name {
			t.Errorf("expected name %s, got %s", f.name, fi.Name)
		}
		if fi.Owner != "test" {
			t.Errorf("expected owner test, got %s", fi.Owner)
		}
		if fi.Path != f.path {
			t.Errorf("expected path %s, got %s", f.path, fi.Path)
		}
		if fi.ReplicaCount() != f.rc {
			t.Errorf("expected ReplicaCount %d, got %d", f.rc, fi.ReplicaCount())
		}
		if fi.Size != f.size {
			t.Errorf("expected size %d, got %d", f.size, fi.Size)
		}
		for k, v := range metaGen(f.meta, create|update) {
			if fi.MetaData(k) != v {
				t.Errorf("%s, meta %s value expected %s, got %s", fi.Name, k, v, fi.MetaData(k))
			}
		}
		for k, v := range metaGen(f.meta, remove) {
			if fi.MetaData(k) == v {
				t.Errorf("%s, meta %s value expected empty, got %s", fi.Name, k, fi.MetaData(k))
			}
		}
	}
}

func TestGetInfoFail(t *testing.T) {
	if _, err := client.GetInfo("%"); err == nil {
		t.Error("expected no error, got ", err)
	}
	if _, err := client.GetInfo("notfound"); err == nil {
		t.Error("expected no error, got ", err)
	}
}

func TestDelFile(t *testing.T) {
	for _, f := range testfiles {
		err := client.Delete(f.path)
		if err != nil {
			t.Fatal(err)
		}
	}
	if err := client.Delete("%"); err == nil {
		t.Error("expected err, got nil")
	}
}

func TestDelDir(t *testing.T) {
	for _, f := range []string{"one", "public/two"} {
		err := client.Delete(f)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestOpenFile(t *testing.T) {
	if _, _, err := OpenFile("nofile", nil); err == nil {
		t.Error("expected error, got <nil>")
	}
	f, err := os.OpenFile(".noread", os.O_CREATE, 0000)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	defer os.Remove(".noread")
	if _, _, err = OpenFile(".noread", nil); err == nil {
		t.Error("expected error, got <nil>")
	}
}
