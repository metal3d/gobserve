package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
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
		log.Println("Killing...", cmd.Process.Pid)
		// cmd.Process.Kill()
		// err := cmd.Process.Signal(syscall.SIGTERM)
		//err := cmd.Process.Release()

		pgid, err := syscall.Getpgid(cmd.Process.Pid)
		if err == nil {
			syscall.Kill(-pgid, 15) // note the minus sign
		}
		cmd.Wait()

		//cmd.Wait()
		cmd = nil
	}

	args := strings.Split(conf.Command, " ")

	log.Printf("Starting: %v", args)

	if len(args) > 1 {

		params := []string{}

		for _, a := range args[1:] {
			l, err := filepath.Glob(a)
			if err != nil {
				log.Fatal(err)
			}
			if len(l) > 0 {
				params = append(params, l...)
			} else {
				params = append(params, a)
			}
		}

		cmd = exec.Command(args[0], params...)
	} else {
		cmd = exec.Command(args[0])
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	err := cmd.Start()

	if err != nil {
		log.Println(err)
	}
	log.Println("PID", cmd.Process.Pid)
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
	files := map[string]bool{}
	for {
		select {
		case <-time.Tick(time.Second * 1):
			for file, _ := range files {
				if !isIgnored(file) {
					go gorun()
					files = make(map[string]bool)
					break
				}
			}
			files = make(map[string]bool)

		case event := <-watcher.Events:
			log.Printf("%+v\n", event)
			files[event.Name] = true
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
