package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v1"

	"github.com/fsnotify/fsnotify"
)

// Conf from yaml file (gobserve.yml).
type Conf struct {
	Watch   []string
	Command string
	Ignore  []string
}

// NewConf returns a *Conf.
func NewConf() *Conf {
	c := Conf{
		Command: "go run *.go",
		Ignore:  []string{"*.*~"},
		Watch:   []string{"."},
	}

	// checking configuration
	if _, err := os.Stat("gobserve.yml"); err == nil {
		log.Println("Reading configuation")

		content, err := ioutil.ReadFile("gobserve.yml")
		if err != nil {
			log.Fatal(err)
		}
		yaml.Unmarshal(content, &c)
	}
	return &c
}

var (
	cmd  *exec.Cmd
	conf = NewConf()
)

// gorun launches command.
func gorun() {

	if cmd != nil {
		log.Println("Killing...")
		cmd.Process.Signal(os.Kill)
		cmd.Wait()
		log.Println("Killed")
		cmd = nil
	}

	args := strings.Split(conf.Command, " ")

	log.Printf("Starting: %v", args)

	if len(args) > 1 {
		cmd = exec.Command(args[0], args[1:]...)
	} else {
		cmd = exec.Command(args[0])
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Println(err)
	}
	cmd.Wait()

}

// isIgnored return true if the file is in ignore list.
func isIgnored(file string) bool {
	counter := 0
	for _, f := range conf.Ignore {
		ok, err := filepath.Match(f, file)
		if err != nil {
			log.Fatal(err)
		}
		if ok {
			counter++
		}
	}
	log.Println(file, "is ignored:", counter > 0)
	return counter > 0
}

// doRefresh calls gorun() after one second when files are changed.
func doRefresh(watcher *fsnotify.Watcher) {
	files := []string{}
	for {
		select {
		case <-time.Tick(time.Second * 1):
			for _, file := range files {
				if !isIgnored(file) {
					go gorun()
					files = []string{}
					break
				}
			}
			files = []string{}

		case event := <-watcher.Events:
			log.Printf("%+v\n", event)
			files = append(files, event.Name)
			log.Println(files)

		case err := <-watcher.Errors:
			log.Fatal(err)
		}
	}
}

func main() {

	log.SetPrefix("GOBSERVE >> ")

	// create a watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	// launch initial startup
	go gorun()
	// start to react on watch events
	go doRefresh(watcher)

	// add working directory
	for _, w := range conf.Watch {
		log.Println("Add watch:", w)
		if err := watcher.Add(w); err != nil {
			log.Fatal("ADDWATCH ERR ", w, err)
		}
	}

	// wait until TERM signal
	<-make(chan bool)
}
