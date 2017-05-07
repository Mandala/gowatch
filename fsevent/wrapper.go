// Copyright (c) 2017 Fadhli Dzil Ikram. All rights reserved.
// Use of source code is governed by a MIT license that can be found in the
// LICENSE file.

// +build darwin

package fsevent

// CGo wrapper that enable FSEvent API access directly from Go.

/*
#cgo LDFLAGS: -framework CoreServices
#include <CoreServices/CoreServices.h>

extern void eventStreamCallback(FSEventStreamRef stream, void *info, size_t numEvents, char **paths,
	FSEventStreamEventFlags *flags, FSEventStreamEventId *ids);
*/
import "C"
import (
	"path/filepath"
	"runtime"
	"unsafe"
)

type fsEventStreamRef C.FSEventStreamRef
type cfRunLoopRef C.CFRunLoopRef

func stringToCFString(str string) C.CFStringRef {
	cstr := C.CString(str)
	defer C.free(unsafe.Pointer(cstr))
	return C.CFStringCreateWithCString(nil, cstr, C.kCFStringEncodingUTF8)
}

//export eventStreamCallback
func eventStreamCallback(stream C.FSEventStreamRef, info unsafe.Pointer, numEvents C.size_t, cpaths **C.char,
	cflags *C.FSEventStreamEventFlags, cids *C.FSEventStreamEventId) {
	// Get stream event from c reference
	es, err := storage.Get(fsEventStreamRef(stream))
	if err != nil {
		return
	}
	// Typecast to Go types
	n := int(numEvents)
	paths := (*[1 << 30]*C.char)(unsafe.Pointer(cpaths))[:n:n]
	flags := (*[1 << 30]C.FSEventStreamEventFlags)(unsafe.Pointer(cflags))[:n:n]
	ids := (*[1 << 30]C.FSEventStreamEventId)(unsafe.Pointer(cids))[:n:n]
	// Forward structured event to event channel
	for i := 0; i < n; i++ {
		es.event <- Event{
			Flags: EventFlags(flags[i]),
			ID:    uint64(ids[i]),
			Path:  C.GoString(paths[i]),
		}
	}
}

func (es *EventStream) Restart() error {
	if es.isRunning() {
		if err := es.Stop(); err != nil {
			return err
		}
	}
	if err := es.Start(); err != nil {
		return err
	}
	return nil
}

func (es *EventStream) Start() error {
	es.mu.Lock()
	defer es.mu.Unlock()
	// Check for runtime status
	if es.stream != nil {
		return ErrInvalidOperation
	}
	// Load default values if unset
	if es.Since == 0 {
		es.Since = EventIDSinceNow
	}
	if es.Latency == 0 {
		es.Latency = 1.0
	}
	if es.event == nil {
		es.event = make(chan Event, 3)
	}
	// Create new CF mutable array for stream event initialization
	cfpaths := C.CFArrayCreateMutable(nil, C.CFIndex(len(es.Path)), &C.kCFTypeArrayCallBacks)
	// Append CF string to CF mutable array
	for _, p := range es.Path {
		path, err := filepath.Abs(p)
		if err != nil {
			return err
		}
		cfpath := stringToCFString(path)
		C.CFArrayAppendValue(cfpaths, unsafe.Pointer(cfpath))
	}
	// Create new event stream
	es.stream = fsEventStreamRef(C.FSEventStreamCreate(nil, (C.FSEventStreamCallback)(C.eventStreamCallback), nil,
		cfpaths, C.FSEventStreamEventId(es.Since), C.CFTimeInterval(es.Latency), C.FSEventStreamCreateFlags(es.Flags)))
	// Clean up CF mutable array
	C.CFRelease(C.CFTypeRef(cfpaths))
	// Add event stream to storage
	storage.Add(es.stream, es)
	// Create synchronization barrier with CFRunLoop thread
	waiter := make(chan struct{})
	// Invoke separate thread for running CFRunLoop
	go func() {
		// Bind goroutine to separate thread
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		es.loop = cfRunLoopRef(C.CFRunLoopGetCurrent())
		C.FSEventStreamScheduleWithRunLoop(es.stream, es.loop, C.kCFRunLoopDefaultMode)
		C.FSEventStreamStart(es.stream)
		close(waiter)
		C.CFRunLoopRun()
	}()
	// Wait until waiter is released and return with no error
	<-waiter
	return nil
}

func (es *EventStream) Stop() error {
	es.mu.Lock()
	defer es.mu.Unlock()
	// Check for runtime status
	if es.stream == nil {
		return ErrInvalidOperation
	}
	// Do event stream cleanup routine
	C.CFRunLoopStop(es.loop)
	C.FSEventStreamStop(es.stream)
	C.FSEventStreamUnscheduleFromRunLoop(es.stream, es.loop, C.kCFRunLoopDefaultMode)
	C.FSEventStreamInvalidate(es.stream)
	C.FSEventStreamRelease(es.stream)
	// Remove event stream from storage and cleanup pointers
	storage.Remove(es.stream)
	es.stream = nil
	es.loop = nil
	close(es.event)
	// Return with no error
	return nil
}
