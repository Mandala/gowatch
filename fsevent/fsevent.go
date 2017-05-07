// Copyright (c) 2017 Fadhli Dzil Ikram. All rights reserved.
// Use of source code is governed by a MIT license that can be found in the
// LICENSE file.

// +build darwin

/*
Package fsevent provide high level access to FSEvent, the native filesystem
events API on Darwin based systems.
*/
package fsevent

import (
	"errors"
	"sync"
)

var ErrInvalidOperation = errors.New("fsevent: Invalid EventStream operation")
var ErrKeyExists = errors.New("fsevent: Storage key already exists")
var ErrKeyNotExists = errors.New("fsevent: Storage key does not exists")

const EventIDSinceNow = uint64((1 << 64) - 1)

type CreateFlags uint32

const (
	CreateFlagUseCFTypes CreateFlags = 1 << iota
	CreateFlagNoDefer
	CreateFlagWatchRoot
	CreateFlagIgnoreSelf
	CreateFlagFileEvents
	CreateFlagMarkSelf
)

type EventFlags uint32

const (
	EventFlagMustScanSubDirs EventFlags = 1 << iota
	EventFlagUserDropped
	EventFlagKernelDropped
	EventFlagIDsWrapped
	EventFlagHistoryDone
	EventFlagRootChanged
	EventFlagMount
	EventFlagUnmount
	EventFlagItemCreated
	EventFlagItemRemoved
	EventFlagItemInodeMetaMod
	EventFlagItemRenamed
	EventFlagItemModified
	EventFlagItemFinderInfoMod
	EventFlagItemChangeOwner
	EventFlagItemXattrMod
	EventFlagItemIsFile
	EventFlagItemIsDir
	EventFlagItemIsSymlink
	EventFlagOwnEvent
	EventFlagItemIsHardlink
	EventFlagItemIsLastHardlink
)

type Event struct {
	ID    uint64
	Path  string
	Flags EventFlags
}

type EventStream struct {
	stream  fsEventStreamRef
	loop    cfRunLoopRef
	mu      sync.Mutex
	event   chan Event
	Path    []string
	Since   uint64
	Latency float64
	Flags   CreateFlags
}

func (es *EventStream) isRunning() bool {
	es.mu.Lock()
	defer es.mu.Unlock()
	return es.stream != nil
}

func (es *EventStream) GetEvent() (Event, error) {
	if !es.isRunning() {
		return Event{}, ErrInvalidOperation
	}
	return <-es.event, nil
}

type eventStreamMap map[fsEventStreamRef]*EventStream

type eventStreamStorage struct {
	storage eventStreamMap
	mu      sync.Mutex
}

func (s *eventStreamStorage) Add(cref fsEventStreamRef, ref *EventStream) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.storage[cref]; ok {
		return ErrKeyExists
	}
	s.storage[cref] = ref
	return nil
}

func (s *eventStreamStorage) Get(cref fsEventStreamRef) (*EventStream, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	es, ok := s.storage[cref]
	if !ok {
		return nil, ErrKeyNotExists
	}
	return es, nil
}

func (s *eventStreamStorage) Remove(cref fsEventStreamRef) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.storage[cref]; !ok {
		return ErrKeyNotExists
	}
	delete(s.storage, cref)
	return nil
}

var storage = eventStreamStorage{storage: make(eventStreamMap)}
