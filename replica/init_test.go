package replica

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

var (
	disks  []string
	userDB string
	token  *Token
	client *Client
	server *exec.Cmd
)

func init() {
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	userDB = filepath.Join(pwd, ".user.db")

	disks = []string{filepath.Join(pwd, ".test1"), filepath.Join(pwd, ".test2")}
	for _, dir := range disks {
		os.Mkdir(dir, 0777)
	}
}

func initServer() {
	f, err := os.Create(userDB)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	_, err = f.WriteString(`{"test":"secret"}`)
	if err != nil {
		log.Fatal(err)
	}
	repPath, err := exec.LookPath("replica")
	if err != nil {
		log.Fatal(err)
	}
	server = exec.Command(repPath,
		"-addr", "localhost:7881", "-user-db", userDB, disks[0], disks[1])
	err = server.Start()
	if err != nil {
		log.Fatal(err)
	}
	time.Sleep(time.Millisecond * 200)
}

func initClient() {
	clt, err := NewClient("")
	if err != nil {
		cleanUp()
		log.Fatal(err)
	}
	token, err = clt.GetToken("test", "secret")
	if err != nil {
		cleanUp()
		log.Fatal(err)
	}
	client, err = NewClient("", AssignToken(token.String()))
	if err != nil {
		cleanUp()
		log.Fatal(err)
	}
}

func cleanUp() {
	err := server.Process.Signal(syscall.SIGTERM)
	if err != nil {
		log.Println(err)
	}
	for _, dir := range disks {
		err := os.RemoveAll(dir)
		if err != nil {
			log.Println(err)
		}
	}
	os.Remove(userDB)
}

func TestCleanUP(t *testing.T) {
	cleanUp()
}
