// Copyright 2014 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package alloter

import (
	"io"
	"net/http"

	"github.com/zhiyunliu/golibs/xtypes"
)

const (
	noWritten     = -1
	defaultStatus = http.StatusOK
)

// ResponseWriter ...
type ResponseWriter interface {
	// Returns the HTTP response status code of the current request.
	Status() int

	// Returns the number of bytes already written into the response http body.
	// See Written()
	Size() int

	// Returns true if the response body was already written.
	Written() bool

	WriteHeader(code int)
	Header() xtypes.SMap
	Write(p []byte) (n int, err error)
	// Writes the string into the response body.
	WriteString(string) (int, error)
	Flush() error
}

type responseWriter struct {
	size   int
	status int
	writer ResponseWriter
}

var _ ResponseWriter = &responseWriter{}

func (w *responseWriter) reset(writer ResponseWriter) {
	w.writer = writer
	w.size = noWritten
	w.status = defaultStatus
}

func (w *responseWriter) WriteHeader(code int) {
	if code > 0 && w.status != code {
		if w.Written() {
			debugPrint("[WARNING] Headers were already written. Wanted to override status code %d with %d", w.status, code)
		}
		w.status = code
	}
}

func (w *responseWriter) Write(data []byte) (n int, err error) {
	n, err = w.writer.Write(data)
	w.size += n
	return
}

func (w *responseWriter) WriteString(s string) (n int, err error) {
	n, err = io.WriteString(w.writer, s)
	w.size += n
	return
}
func (w *responseWriter) Header() xtypes.SMap {
	return w.writer.Header()
}

func (w *responseWriter) Status() int {
	return w.status
}

func (w *responseWriter) Size() int {
	return w.size
}

func (w *responseWriter) Written() bool {
	return w.size != noWritten
}

// Flush implements the http.Flush interface.
func (w *responseWriter) Flush() error {
	return w.writer.Flush()
}
