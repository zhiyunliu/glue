package alloter

import (
	"encoding/json"
	"encoding/xml"
	"fmt"

	"github.com/zhiyunliu/golibs/bytesconv"
)

func writeContentType(w ResponseWriter, value string) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = value
	}
}

// Render interface is to be implemented by JSON, XML, HTML, YAML and so on.
type Render interface {
	// Render writes data with custom ContentType.
	Render(ResponseWriter) error
	// WriteContentType writes custom ContentType.
	WriteContentType(w ResponseWriter)
}

// Data contains ContentType and bytes data.
type DataRender struct {
	ContentType string
	Data        []byte
}

// Render (Data) writes data with custom ContentType.
func (r DataRender) Render(w ResponseWriter) (err error) {
	r.WriteContentType(w)
	_, err = w.Write(r.Data)
	return
}

// WriteContentType (Data) writes custom ContentType.
func (r DataRender) WriteContentType(w ResponseWriter) {
	writeContentType(w, r.ContentType)
}

// String contains the given interface object slice and its format.
type StringRender struct {
	Format string
	Data   []interface{}
}

var plainContentType = "text/plain; charset=utf-8"

// Render (String) writes data with custom ContentType.
func (r StringRender) Render(w ResponseWriter) error {
	return writeString(w, r.Format, r.Data)
}

// WriteContentType (String) writes Plain ContentType.
func (r StringRender) WriteContentType(w ResponseWriter) {
	writeContentType(w, plainContentType)
}

// WriteString writes data according to its format and write custom ContentType.
func writeString(w ResponseWriter, format string, data []interface{}) (err error) {
	writeContentType(w, plainContentType)
	if len(data) > 0 {
		_, err = fmt.Fprintf(w, format, data...)
		return
	}
	_, err = w.Write(bytesconv.StringToBytes(format))
	return
}

// XML contains the given interface object.
type XMLRender struct {
	Data interface{}
}

var xmlContentType = "application/xml; charset=utf-8"

// Render (XML) encodes the given interface object and writes data with custom ContentType.
func (r XMLRender) Render(w ResponseWriter) error {
	r.WriteContentType(w)
	return xml.NewEncoder(w).Encode(r.Data)
}

// WriteContentType (XML) writes XML ContentType for response.
func (r XMLRender) WriteContentType(w ResponseWriter) {
	writeContentType(w, xmlContentType)
}

// JSON contains the given interface object.
type JSONRender struct {
	Data interface{}
}

var jsonContentType = "application/json; charset=utf-8"

// Render (JSON) writes data with custom ContentType.
func (r JSONRender) Render(w ResponseWriter) (err error) {
	if err = WriteJSON(w, r.Data); err != nil {
		panic(err)
	}
	return
}

// WriteContentType (JSON) writes JSON ContentType.
func (r JSONRender) WriteContentType(w ResponseWriter) {
	writeContentType(w, jsonContentType)
}

// WriteJSON marshals the given interface object and writes it with custom ContentType.
func WriteJSON(w ResponseWriter, obj interface{}) error {
	writeContentType(w, jsonContentType)
	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	_, err = w.Write(jsonBytes)
	return err
}
