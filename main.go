package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"path/filepath"
	"strconv"
	"time"

	"sync"

	"strings"

	"github.com/mandala/gowatch/wtch"
)

type locker struct {
	state bool
	Sleep time.Duration
	mu    sync.Mutex
}

func (l *locker) Get() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return !l.state
}

func (l *locker) Set() {
	go func() {
		time.Sleep(l.Sleep)
		l.mu.Lock()
		defer l.mu.Unlock()
		l.state = false
	}()
}

var taskPath string
var task *exec.Cmd

func watchHandler(w *wtch.Watcher) {
	l := locker{
		Sleep: 2 * time.Second,
	}
	for {
		// Get watch event
		ev, err := w.Get()
		if err != nil {
			fmt.Printf("ERROR - GetEvent, %s.\n", err.Error())
		}
		// Ignore event on lock
		if !l.Get() {
			continue
		}
		// Check event
		if !ev.IsDir() && filepath.Ext(ev.Path) == ".go" {
			l.Set()
			rebuildTask()
		} else if strings.HasPrefix(ev.Path, "resources/") {
			l.Set()
			restartTask()
		}
	}
}

func stopTask() {
	if task != nil {
		task.Process.Kill()
	}
	task = nil
}

func startTask() {
	fmt.Println("INFO  - Starting application.")
	t := exec.Command(taskPath, os.Args[1:]...)
	t.Stderr = os.Stderr
	t.Stdout = os.Stdout
	if err := t.Start(); err != nil {
		fmt.Printf("ERROR - StartTask, %s.\n", err.Error())
		return
	}
	task = t
}

func cleanTask() {
	stopTask()
	if taskPath != "" {
		if _, err := os.Stat(taskPath); err == nil {
			os.Remove(taskPath)
		}
	}
	taskPath = ""
}

func rebuildTask() {
	cleanTask()
	fmt.Println("INFO  - Rebuilding project.")
	// create new tmpfile
	taskPath = filepath.Join(os.TempDir(), "gowatch-build-"+strconv.Itoa(int(time.Now().Unix())))
	// Start compiler
	c := exec.Command("go", "build", "-o", taskPath)
	c.Stderr = os.Stderr
	c.Stdout = os.Stdout
	if err := c.Run(); err != nil {
		fmt.Printf("ERROR - RebuildTask, %s.\n", err.Error())
	}
	// Start task
	startTask()
}

func restartTask() {
	stopTask()
	startTask()
}

func main() {
	// Create new watcher
	watcher, err := wtch.NewWatcher("")
	if err != nil {
		fmt.Printf("ERROR - NewWatcher, %s.\n", err.Error())
		os.Exit(1)
	}
	// Invoke task rebuilding
	rebuildTask()
	// Start the watcher
	fmt.Printf("INFO  - Starting watcher.\n")
	watcher.Start()
	// Run watch handler in separate goroutine
	go watchHandler(watcher)
	// Wait until interrupt
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	<-sig
	fmt.Printf(" --- interrupt ---\n")
	// Do watcher cleanup operation
	fmt.Printf("INFO  - Terminating watcher.\n")
	watcher.Stop()
}
