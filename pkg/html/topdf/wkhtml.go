package topdf

import (
	"bytes"
	"context"
	"io"
	"kafka-polygon/pkg/cerror"
)

const (
	_defaultDPI = 300
)

type WHConvertor interface {
	WithContext(ctx context.Context)
	SetDPI(dpi uint)
	SetPageSize(ps string)
	SetHTML(h []byte)
	ToPDF() (*bytes.Buffer, error)
	ToWritePDF(w io.Writer) (err error)
	ToSavePDF(filename string) error
	ResetPages()
	Buffer() *bytes.Buffer
}

type Convertor struct {
	ctx       context.Context
	dpi       uint
	pageSize  string
	html      []byte
	generator *wkhtmltopdf.PDFGenerator
}

func New() (WHConvertor, error) {
	bgCtx := context.Background()
	gen, err := wkhtmltopdf.NewPDFGenerator()

	if err != nil {
		return nil, cerror.NewF(bgCtx, cerror.KindInternal, "wkhtmltopdf init err: %s", err).
			LogError()
	}

	return &Convertor{
		dpi:       _defaultDPI,
		pageSize:  wkhtmltopdf.PageSizeA4,
		generator: gen,
		ctx:       bgCtx,
	}, nil
}

func (c *Convertor) SetDPI(dpi uint) {
	c.dpi = dpi
}

func (c *Convertor) SetPageSize(ps string) {
	c.pageSize = ps
}

func (c *Convertor) SetHTML(h []byte) {
	c.html = h
}

func (c *Convertor) WithContext(ctx context.Context) {
	c.ctx = ctx
}

func (c *Convertor) ToPDF() (*bytes.Buffer, error) {
	err := c.create()
	if err != nil {
		return nil, cerror.NewF(c.ctx, cerror.KindInternal, "wkhtmltopdf ToPdf::create: %s", err).
			LogError()
	}

	return c.generator.Buffer(), nil
}

func (c *Convertor) ToWritePDF(w io.Writer) (err error) {
	err = c.create()

	if err != nil {
		return cerror.NewF(c.ctx, cerror.KindInternal, "wkhtmltopdf ToWritePdf::create: %s", err).
			LogError()
	}

	c.generator.SetOutput(w)

	return
}

func (c *Convertor) ToSavePDF(filename string) error {
	err := c.create()

	if err != nil {
		return cerror.NewF(c.ctx, cerror.KindInternal, "wkhtmltopdf ToSavePdf::create: %s", err).
			LogError()
	}

	err = c.generator.WriteFile(filename)
	if err != nil {
		return cerror.NewF(c.ctx, cerror.KindInternal, "wkhtmltopdf ToSavePdf::WriteFile: %s", err).
			LogError()
	}

	return nil
}

func (c *Convertor) create() error {
	c.generator.AddPage(wkhtmltopdf.NewPageReader(bytes.NewReader(c.html)))
	c.generator.PageSize.Set(c.pageSize)
	c.generator.Dpi.Set(c.dpi)

	return c.generator.CreateContext(c.ctx)
}

func (c *Convertor) ResetPages() {
	c.generator.ResetPages()
}

func (c *Convertor) Buffer() *bytes.Buffer {
	return c.generator.Buffer()
}
