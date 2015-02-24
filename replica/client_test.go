package replica

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	addrs := map[string]string{
		"":                       "http://localhost:7881/json",
		"addr":                   "http://addr:7881/json",
		"http://addr":            "http://addr:7881/json",
		"https://addr":           "https://addr:7881/json",
		"addr:8808":              "http://addr:8808/json",
		"https://addr:8808/json": "https://addr:8808/json",
		"addr/redirect":          "http://addr/redirect/json",
		"addr:9192/redirect":     "http://addr:9192/redirect/json",
	}

	for addr, exres := range addrs {
		clt, err := NewClient(addr)
		if err != nil {
			t.Error(err)
		}
		if clt.Address() != exres {
			t.Errorf("expected %s, got %s", exres, clt.addr)
		}
	}
}

func TestNewClientFuncs(t *testing.T) {
	clt, err := NewClient("", AllowUnsignedSSL, AssignToken("123456789"))
	if err != nil {
		t.Fatal(err)
	}
	if !clt.unsecureSSL {
		t.Error("AllowUnsignedSSL failed")
	}
	if clt.token.Token != "123456789" {
		t.Error("AssigneToken failed")
	}
}

func TestClientFail(t *testing.T) {
	_, err := NewClient("%")
	if err == nil {
		t.Error("expected url parse error, got <nil>")
	}

	clt, err := NewClient("https://localhost")
	if err != nil {
		t.Fatal(err)
	}
	_, err = clt.GetToken("", "")
	if err == nil {
		t.Error("expected connection error, got <nil>")
	}

	_, err = clt.newRequest("NA", "%", nil)
	if err == nil {
		t.Error("expected error, got <nil>")
	}
}

func TestClientJoinURL(t *testing.T) {
	clt, _ := NewClient("")
	for _, name := range []string{"test", "/test/", "/test", "test/"} {
		res := clt.joinURL(name)
		if res != "http://localhost:7881/json/test" {
			t.Error("expected 'http://localhost:7881/json/test', got ", res)
		}
	}
}

func TestGetTokenFail(t *testing.T) {
	clt, _ := NewClient("")
	clt.addr = "%"
	if _, err := clt.GetToken("", ""); err == nil {
		t.Error("expected error, got <nil>")
	}
	if _, err := clt.Token(); err == nil {
		t.Error("expected token not set error, got <nil>")
	}
	//json decode error
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("ContentType", "application/json")
		w.WriteHeader(200)
		fmt.Fprintln(w, "l")
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()

	clt, _ = NewClient(ts.URL)
	_, err := clt.GetToken("", "")
	if err == nil {
		t.Error("expected json decode error, got <nil>")
	}
}

type failedReadCloser struct{}

func (f *failedReadCloser) Read(p []byte) (int, error) { return 0, fmt.Errorf("read error") }
func (f *failedReadCloser) Close() error               { return nil }

func TestParseResponseFail(t *testing.T) {
	resp := new(http.Response)
	resp.StatusCode = 404
	resp.Body = new(failedReadCloser)
	err := parsResponse(resp)
	if err == nil {
		t.Error("expected read error, got <nil>")
	}
}
