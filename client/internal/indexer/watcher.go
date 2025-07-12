package indexer

import (
	"log"

	"github.com/fsnotify/fsnotify"
)

func NewWatcher(paths []string, onEvent func(event fsnotify.Event)) (*fsnotify.Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	for _, path := range paths {
		err = watcher.Add(path)
		if err != nil {
			watcher.Close()
			return nil, err
		}
	}
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				onEvent(event)
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("Watcher error:", err)
			}
		}
	}()
	return watcher, nil
}
