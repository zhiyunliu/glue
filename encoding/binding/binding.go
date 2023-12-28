// Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

//go:build !nomsgpack

package binding

import (
	"github.com/zhiyunliu/glue/encoding"
)

// Content-Type MIME of the most common data formats.
const (
	MIMEJSON              = "application/json"
	MIMEHTML              = "text/html"
	MIMEXML               = "application/xml"
	MIMEXML2              = "text/xml"
	MIMEPlain             = "text/plain"
	MIMEPOSTForm          = "application/x-www-form-urlencoded"
	MIMEMultipartPOSTForm = "multipart/form-data"
	MIMEPROTOBUF          = "application/x-protobuf"
	MIMEYAML              = "application/x-yaml"
	MIMETOML              = "application/toml"
)

// // These implement the Binding interface and can be used to bind the data
// // present in the request to struct instances.
// var (
// 	JSON     = jsonBinding{}
// 	XML      = xmlBinding{}
// 	Form     = formBinding{}
// 	FormPost = formPostBinding{}
// 	//FormMultipart = formMultipartBinding{}
// 	ProtoBuf = protobufBinding{}
// 	YAML     = yamlBinding{}
// 	TOML     = tomlBinding{}
// )

// // Default returns the appropriate Binding instance based on the HTTP method
// // and the content type.
// func Default(method, contentType string) encoding.Codec {
// 	if method == http.MethodGet {
// 		return Form
// 	}

// 	switch contentType {
// 	case MIMEJSON:
// 		return JSON
// 	case MIMEXML, MIMEXML2:
// 		return XML
// 	case MIMEPROTOBUF:
// 		return ProtoBuf
// 	case MIMEYAML:
// 		return YAML
// 	case MIMETOML:
// 		return TOML
// 	// case MIMEMultipartPOSTForm:
// 	// 	return FormMultipart
// 	default: // case MIMEPOSTForm:
// 		return Form
// 	}
// }

func init() {
	encoding.RegisterCodec(jsonBinding{})
	encoding.RegisterCodec(xmlBinding{})
	encoding.RegisterCodec(formBinding{})
	encoding.RegisterCodec(formPostBinding{})
	encoding.RegisterCodec(protobufBinding{})
	encoding.RegisterCodec(xprotobufBinding{})
	encoding.RegisterCodec(yamlBinding{})
	encoding.RegisterCodec(xyamlBinding{})
	encoding.RegisterCodec(tomlBinding{})
}
