package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"slices"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
)

func serveIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	f, err := os.ReadFile("./index.html")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprint(w, string(f))
}
func serveWs(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case <-watchChan:
			b2 := "reload"
			ws.WriteMessage(websocket.TextMessage, []byte(b2))
		}
	}
}
func handleWatcher(w *fsnotify.Watcher, wChan chan int) {
	for {
		select {
		case event, ok := <-w.Events:
			if !ok {
				log.Fatal("Event not ok")
			}
			switch event.Op {
			case fsnotify.Create:
				handleCreate(w, event)
				break
			case fsnotify.Remove:
				handleRemove(w, event)
				break
			}
			wChan <- 1
		case error, ok := <-w.Errors:
			if !ok {
				log.Fatal("Errors not ok")
			}
			log.Fatal(error)
		}

	}
}

func handleRemove(w *fsnotify.Watcher, event fsnotify.Event) {
	if slices.Contains(w.WatchList(), event.Name) {
		w.Remove(event.Name)
	}
}

func handleCreate(w *fsnotify.Watcher, event fsnotify.Event) {
	fi, err := os.Stat(event.Name)
	if err != nil {
		log.Println("ERROR[handleCreate]: ", err)
		return
	}
	if !fi.IsDir() {
		return
	}
	if !slices.Contains(w.WatchList(), event.Name) {
		w.Add(event.Name)
	}
}

var watchChan = make(chan int, 1)

func main() {
	watcher, err := fsnotify.NewWatcher()
	pwd, err := filepath.Abs("./")
	if err != nil {
		log.Fatal(err)
	}
	watcher.Add(pwd)
	defer watcher.Close()

	go handleWatcher(watcher, watchChan)

	http.HandleFunc("/", serveIndex)
	http.HandleFunc("/ws", serveWs)

	if err := http.ListenAndServe("localhost:8080", nil); err != nil {
		log.Fatal(err)
	}
}
