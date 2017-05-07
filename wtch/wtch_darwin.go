// Copyright (c) 2017 Fadhli Dzil Ikram. All rights reserved.
// Use of source code is governed by a MIT license that can be found in the
// LICENSE file.

// +build darwin

package wtch

import "github.com/mandala/gowatch/fsevent"
import "path/filepath"

type Watcher struct {
	es   fsevent.EventStream
	base string
}

func (w *Watcher) Get() (Event, error) {
	// Get events from FSEvents
	fse, err := w.es.GetEvent()
	if err != nil {
		return Event{}, err
	}
	// Convert to commonspec
	rel, err := filepath.Rel(w.base, fse.Path)
	if err != nil {
		return Event{}, err
	}
	var flags EventFlags
	flags |= EventFlags(fse.Flags>>8) & EventFlagItemCreated
	flags |= EventFlags(fse.Flags>>11) & EventFlagItemModified
	flags |= EventFlags(fse.Flags>>9) & EventFlagItemRenamed
	flags |= EventFlags(fse.Flags>>6) & EventFlagItemRemoved
	flags |= EventFlags(fse.Flags>>13) & EventFlagItemIsDir
	// Return as event object
	return Event{
		Path:  rel,
		Flags: flags,
	}, nil
}

func (w *Watcher) Start() error {
	return w.es.Start()
}

func (w *Watcher) Stop() error {
	return w.es.Stop()
}

func newWatcher(base string) (*Watcher, error) {
	// Initialize new watcher object
	return &Watcher{
		es: fsevent.EventStream{
			Flags:   fsevent.CreateFlagFileEvents,
			Latency: 0.1,
			Path:    []string{base},
		},
		base: base,
	}, nil
}
