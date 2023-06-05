package topdf_test

import (
	"kafka-polygon/pkg/html/topdf"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tj/assert"
)

const (
	html = `<!doctype html><html><head><title>WKHTMLTOPDF TEST</title></head><body>HELLO PDF</body></html>`
)

func TestConvertor(t *testing.T) {
	t.Parallel()

	path := "./testdata/simplesample1.pdf"

	defer func() {
		err := os.Remove(path)
		require.NoError(t, err)
	}()

	conv, err := topdf.New()
	require.NoError(t, err)

	conv.SetHTML([]byte(html))
	err = conv.ToSavePDF(path)
	require.NoError(t, err)

	defer conv.Buffer().Reset()

	conv.ResetPages()

	conv.SetHTML([]byte(`<b>test bold</b>`))
	buff, err := conv.ToPDF()
	require.NoError(t, err)

	defer conv.Buffer().Reset()

	assert.True(t, buff.Len() > 0)
}

func TestConvertorToPdf(t *testing.T) {
	t.Parallel()

	conv, err := topdf.New()
	require.NoError(t, err)

	conv.SetHTML([]byte(html))
	conv.SetDPI(600)
	conv.SetPageSize(wkhtmltopdf.PageSizeA5)
	buff, err := conv.ToPDF()
	require.NoError(t, err)

	defer conv.Buffer().Reset()

	assert.True(t, buff.Len() > 0)
}

func TestCheckPDF(t *testing.T) {
	t.Parallel()

	path := "./testdata/simplesample.pdf"

	f, err := os.Open(path)
	require.NoError(t, err)

	fi, err := f.Stat()
	require.NoError(t, err)

	defer f.Close()

	assert.True(t, fi.Size() > 0)
	assert.Equal(t, fi.Size(), int64(5247))
}
