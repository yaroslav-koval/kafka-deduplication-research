package openapi

import (
	"encoding/xml"
	"io"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
)

type CustomBodyDecoder map[string]openapi3filter.BodyDecoder

func NewCustomBodyDecoders() CustomBodyDecoder {
	return CustomBodyDecoder{
		"text/xml": plainXMLBodyDecoder,
	}
}

func plainXMLBodyDecoder(body io.Reader, header http.Header, schema *openapi3.SchemaRef, encFn openapi3filter.EncodingFn) (
	interface{}, error) {
	var value interface{}
	if err := xml.NewDecoder(body).Decode(&value); err != nil {
		return nil, &openapi3filter.ParseError{Kind: openapi3filter.KindInvalidFormat, Cause: err}
	}

	return value, nil
}
