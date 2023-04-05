package inotify

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/sys/unix"
)

type Inotify struct {
	fd          int
	inotifyFile *os.File
	paths       map[int]string
	locker      sync.Mutex
	Event       chan InotifyEvent
	Error       chan error
	isClose     chan struct{}
	gorutine    chan struct{}
}

type InotifyEvent struct {
	Name      string
	Operation Operation
}

type Operation uint32

const (
	Create Operation = 1 << iota
	Write
	Remove
	Rename
	Chmod
)

func NewINotify() (*Inotify, error) {
	fd, err := unix.InotifyInit1(0)
	if fd == -1 {
		return nil, err
	}

	i := &Inotify{
		fd:          fd,
		paths:       make(map[int]string),
		inotifyFile: os.NewFile(uintptr(fd), ""),
		Event:       make(chan InotifyEvent),
		Error:       make(chan error),
	}

	go i.readEvent()
	return i, nil
}

func (i *Inotify) Add(name string) error {
	name = filepath.Clean(name)
	watchDesc, err := unix.InotifyAddWatch(i.fd, name, unix.IN_MOVED_TO|unix.IN_MOVED_FROM|unix.IN_CREATE|unix.IN_ATTRIB|unix.IN_MODIFY|unix.IN_MOVE_SELF|unix.IN_DELETE|unix.IN_DELETE_SELF)
	if err != nil {
		return err
	}

	i.locker.Lock()
	i.paths[watchDesc] = name
	i.locker.Unlock()

	return nil
}

func (i *Inotify) Remove(name string) error {
	name = filepath.Clean(name)

	for k, v := range i.paths {
		if v == name {
			i.locker.Lock()
			delete(i.paths, k)
			i.locker.Unlock()

			_, err := unix.InotifyRmWatch(i.fd, uint32(k))
			return err
		}
	}
	return fmt.Errorf("unkown inotify file: %s", name)
}

func (i *Inotify) List() []string {
	desc := []string{}
	for _, v := range i.paths {
		desc = append(desc, v)
	}
	return desc
}

func (i *Inotify) Close() error {
	if i.isClosed() {
		return nil
	}

	close(i.isClose)
	err := i.inotifyFile.Close()
	if err != nil {
		return nil
	}

	<-i.gorutine

	return nil
}

func (i *Inotify) readEvent() {
	defer func() {
		close(i.Event)
		close(i.Error)
		close(i.gorutine)
	}()

	var buf [unix.SizeofInotifyEvent * 4096]byte

	for {
		n, err := i.inotifyFile.Read(buf[:])
		switch {
		case errors.Unwrap(err) == os.ErrClosed:
			return
		case err != nil:
			if !i.sendError(err) {
				return
			}
			continue
		}

		if n < unix.SizeofInotifyEvent {
			var err error
			if n == 0 {
				err = io.EOF
			} else if n < 0 {
				err = io.ErrUnexpectedEOF
			} else {
				err = io.ErrShortBuffer
			}
			if !i.sendError(err) {
				return
			}
			continue
		}

		var (
			event  = unix.InotifyEvent{}
			reader = bytes.NewReader(buf[:n])
			offset = 0
		)
		for offset+unix.SizeofInotifyEvent < n {
			// read Event
			binary.Read(reader, binary.LittleEndian, &event)
			name := make([]byte, event.Len)
			// read Name by Event
			binary.Read(reader, binary.LittleEndian, &name)

			// Check "Delete" Event
			root, ok := i.paths[int(event.Wd)]
			if ok && event.Mask&unix.IN_DELETE_SELF == unix.IN_DELETE_SELF {
				i.locker.Lock()
				delete(i.paths, int(event.Wd))
				i.locker.Unlock()
			}

			// Create Event
			var sendEvent InotifyEvent
			//// Set Event Operation
			if event.Mask&unix.IN_CREATE == unix.IN_CREATE || event.Mask&unix.IN_MOVED_TO == unix.IN_MOVED_TO {
				sendEvent.Operation |= Create
			}
			if event.Mask&unix.IN_DELETE_SELF == unix.IN_DELETE_SELF || event.Mask&unix.IN_DELETE == unix.IN_DELETE {
				sendEvent.Operation |= Remove
			}
			if event.Mask&unix.IN_MODIFY == unix.IN_MODIFY {
				sendEvent.Operation |= Write
			}
			if event.Mask&unix.IN_MOVE_SELF == unix.IN_MOVE_SELF || event.Mask&unix.IN_MOVED_FROM == unix.IN_MOVED_FROM {
				sendEvent.Operation |= Rename
			}
			if event.Mask&unix.IN_ATTRIB == unix.IN_ATTRIB {
				sendEvent.Operation |= Chmod
			}
			//// Set Event FileName
			sendEvent.Name = string(name)
			if ok {
				sendEvent.Name = path.Join(root, string(name))
			}

			// Send Event Check
			if event.Mask&unix.IN_IGNORED != unix.IN_IGNORED {
				if !i.sendEvent(sendEvent) {
					return
				}
			}

			// Seek to Next Event
			offset = offset + unix.SizeofInotifyEvent + int(event.Len)
		}
	}
}

func (i *Inotify) isClosed() bool {
	select {
	case <-i.isClose:
		return true
	default:
		return false
	}
}

func (i *Inotify) sendError(err error) (success bool) {
	select {
	case i.Error <- err:
		return true
	case <-i.isClose:
		return false
	}
}

func (i *Inotify) sendEvent(event InotifyEvent) (success bool) {
	select {
	case i.Event <- event:
		return true
	case <-i.isClose:
		return false
	}
}

func (e InotifyEvent) Has(o Operation) bool { return e.Operation.Has(o) }
func (o Operation) Has(h Operation) bool    { return o&h == h }

func (o Operation) ToString() string {
	var b strings.Builder
	if o.Has(Create) {
		b.WriteString("|CREATE")
	}
	if o.Has(Remove) {
		b.WriteString("|REMOVE")
	}
	if o.Has(Write) {
		b.WriteString("|WRITE")
	}
	if o.Has(Rename) {
		b.WriteString("|RENAME")
	}
	if o.Has(Chmod) {
		b.WriteString("|CHMOD")
	}
	if b.Len() == 0 {
		return ""
	}
	return b.String()[1:]
}
