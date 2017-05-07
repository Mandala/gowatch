// Copyright (c) 2017 Fadhli Dzil Ikram. All rights reserved.
// Use of source code is governed by a MIT license that can be found in the
// LICENSE file.

/*
Package wtch provide common file event watcher API endpoints across different
platforms.
*/
package wtch

import (
	"os"
	"path/filepath"
)

type EventFlags uint32

const (
	EventFlagItemCreated EventFlags = 1 << iota
	EventFlagItemModified
	EventFlagItemRenamed
	EventFlagItemRemoved
	EventFlagItemIsDir
)

type Event struct {
	Path  string
	Flags EventFlags
}

func (e Event) IsDir() bool {
	return (e.Flags & EventFlagItemIsDir) != 0
}

func NewWatcher(base string) (*Watcher, error) {
	// Get absolute path from base
	p, err := filepath.Abs(base)
	if err != nil {
		return nil, err
	}
	// Checks if path is really exists
	if _, err := os.Stat(p); err != nil {
		return nil, err
	}
	// Return new watcher object from platform-specific initialization
	return newWatcher(p)
}
