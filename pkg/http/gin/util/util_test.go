package util_test

import (
	"kafka-polygon/pkg/http/gin/util"
	"mime"
	"net/http"
	"testing"

	"github.com/tj/assert"
)

func Test_GetMIME(t *testing.T) {
	t.Parallel()

	res := util.GetMIME(".json")
	assert.Equal(t, "application/json", res)

	res = util.GetMIME(".xml")
	assert.Equal(t, "application/xml", res)

	res = util.GetMIME("xml")
	assert.Equal(t, "application/xml", res)

	res = util.GetMIME("unknown")
	assert.Equal(t, util.MIMEOctetStream, res)
	// empty case
	res = util.GetMIME("")
	assert.Equal(t, "", res)
}

// go test -v -run=^$ -bench=Benchmark_GetMIME -benchmem -count=2
func Benchmark_GetMIME(b *testing.B) {
	var res string

	b.Run("fiber", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			res = util.GetMIME(".xml")
			res = util.GetMIME(".txt")
			res = util.GetMIME(".png")
			res = util.GetMIME(".exe")
			res = util.GetMIME(".json")
		}
		assert.Equal(b, "application/json", res)
	})
	b.Run("default", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			res = mime.TypeByExtension(".xml")
			res = mime.TypeByExtension(".txt")
			res = mime.TypeByExtension(".png")
			res = mime.TypeByExtension(".exe")
			res = mime.TypeByExtension(".json")
		}
		assert.Equal(b, "application/json", res)
	})
}

func Test_ParseVendorSpecificContentType(t *testing.T) {
	t.Parallel()

	cType := util.ParseVendorSpecificContentType("application/json")
	assert.Equal(t, "application/json", cType)

	cType = util.ParseVendorSpecificContentType(
		"multipart/form-data; boundary=dart-http-boundary-ZnVy.ICWq+7HOdsHqWxCFa8g3D.KAhy+Y0sYJ_lBADypu8po3_X")
	assert.Equal(t, "multipart/form-data", cType)

	cType = util.ParseVendorSpecificContentType("multipart/form-data")
	assert.Equal(t, "multipart/form-data", cType)

	cType = util.ParseVendorSpecificContentType("application/vnd.api+json; version=1")
	assert.Equal(t, "application/json", cType)

	cType = util.ParseVendorSpecificContentType("application/vnd.api+json")
	assert.Equal(t, "application/json", cType)

	cType = util.ParseVendorSpecificContentType("application/vnd.dummy+x-www-form-urlencoded")
	assert.Equal(t, "application/x-www-form-urlencoded", cType)

	cType = util.ParseVendorSpecificContentType("something invalid")
	assert.Equal(t, "something invalid", cType)
}

func Benchmark_ParseVendorSpecificContentType(b *testing.B) {
	var cType string

	b.Run("vendorContentType", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			cType = util.ParseVendorSpecificContentType("application/vnd.api+json; version=1")
		}
		assert.Equal(b, "application/json", cType)
	})

	b.Run("defaultContentType", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			cType = util.ParseVendorSpecificContentType("application/json")
		}
		assert.Equal(b, "application/json", cType)
	})
}

func Test_StatusMessage(t *testing.T) {
	t.Parallel()

	res := util.StatusMessage(204)
	assert.Equal(t, "No Content", res)

	res = util.StatusMessage(404)
	assert.Equal(t, "Not Found", res)

	res = util.StatusMessage(426)
	assert.Equal(t, "Upgrade Required", res)

	res = util.StatusMessage(511)
	assert.Equal(t, "Network Authentication Required", res)

	res = util.StatusMessage(1337)
	assert.Equal(t, "", res)

	res = util.StatusMessage(-1)
	assert.Equal(t, "", res)

	res = util.StatusMessage(0)
	assert.Equal(t, "", res)

	res = util.StatusMessage(600)
	assert.Equal(t, "", res)
}

// go test -run=^$ -bench=Benchmark_StatusMessage -benchmem -count=2
func Benchmark_StatusMessage(b *testing.B) {
	var res string

	b.Run("fiber", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			res = util.StatusMessage(http.StatusNotExtended)
		}
		assert.Equal(b, "Not Extended", res)
	})
	b.Run("default", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			res = http.StatusText(http.StatusNotExtended)
		}
		assert.Equal(b, "Not Extended", res)
	})
}
