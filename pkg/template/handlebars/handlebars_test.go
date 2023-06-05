package handlebars_test

import (
	"bytes"
	"html"
	"kafka-polygon/pkg/template/handlebars"
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/tj/assert"
)

func trim(str string) string {
	trimmed := strings.TrimSpace(regexp.MustCompile(`\s+`).ReplaceAllString(str, " "))
	trimmed = strings.ReplaceAll(trimmed, " <", "<")
	trimmed = strings.ReplaceAll(trimmed, "> ", ">")

	return trimmed
}

func Test_Render(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	engine := handlebars.New("./views", ".hbs")
	if err := engine.Load(); err != nil {
		t.Fatalf("load: %v\n", err)
	}

	_ = engine.Render(&buf, "index", map[string]interface{}{
		"Title": "Hello, World!",
	})

	expect := `<h2>Header</h2><h1>Hello, World!</h1><h2>Footer</h2>`
	result := trim(buf.String())

	if expect != result {
		t.Fatalf("Expected:\n%s\nResult:\n%s\n", expect, result)
	}

	buf.Reset()

	_ = engine.Render(&buf, "errors/404", map[string]interface{}{
		"Title": "Hello, World!",
	})

	expect = `<h1>Hello, World!</h1>`

	result = trim(buf.String())
	if expect != result {
		t.Fatalf("Expected:\n%s\nResult:\n%s\n", expect, result)
	}
}

func Test_Layout(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	engine := handlebars.New("./views", ".hbs")
	if err := engine.Load(); err != nil {
		t.Fatalf("load: %v\n", err)
	}

	_ = engine.Render(&buf, "index", map[string]interface{}{
		"Title": "Hello, World!",
	}, "layouts/main")

	expect := `<!DOCTYPE html><html><head><title>Main</title></head><body><h2>Header</h2><h1>Hello, World!</h1><h2>Footer</h2></body></html>`

	result := trim(buf.String())
	if expect != result {
		t.Fatalf("Expected:\n%s\nResult:\n%s\n", expect, result)
	}
}

func Test_Empty_Layout(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	engine := handlebars.New("./views", ".hbs")
	if err := engine.Load(); err != nil {
		t.Fatalf("load: %v\n", err)
	}

	_ = engine.Render(&buf, "index", map[string]interface{}{
		"Title": "Hello, World!",
	}, "")

	expect := `<h2>Header</h2><h1>Hello, World!</h1><h2>Footer</h2>`

	result := trim(buf.String())
	if expect != result {
		t.Fatalf("Expected:\n%s\nResult:\n%s\n", expect, result)
	}
}

func Test_FileSystem(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	engine := handlebars.NewFileSystem(http.Dir("./views"), ".hbs")
	engine.Debug(true)

	if err := engine.Load(); err != nil {
		t.Fatalf("load: %v\n", err)
	}

	_ = engine.Render(&buf, "index", map[string]interface{}{
		"Title": "Hello, World!",
	}, "layouts/main")

	expect := `<!DOCTYPE html><html><head><title>Main</title></head><body><h2>Header</h2><h1>Hello, World!</h1><h2>Footer</h2></body></html>`

	result := trim(buf.String())
	if expect != result {
		t.Fatalf("Expected:\n%s\nResult:\n%s\n", expect, result)
	}
}

func Test_Reload(t *testing.T) {
	var buf bytes.Buffer

	engine := handlebars.NewFileSystem(http.Dir("./views"), ".hbs")
	engine.Reload(true) // Optional. Default: false

	if err := engine.Load(); err != nil {
		t.Fatalf("load: %v\n", err)
	}

	if err := os.WriteFile("./views/reload.hbs", []byte("after reload\n"), 0600); err != nil {
		t.Fatalf("write file: %v\n", err)
	}

	defer func() {
		if err := os.WriteFile("./views/reload.hbs", []byte("before reload\n"), 0600); err != nil {
			t.Fatalf("write file: %v\n", err)
		}
	}()

	_ = engine.Load()

	_ = engine.Render(&buf, "reload", nil)

	expect := "after reload"
	result := trim(buf.String())

	if expect != result {
		t.Fatalf("Expected:\n%s\nResult:\n%s\n", expect, result)
	}
}

func TestCustomIfHelper(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	engine := handlebars.New("./views", ".hbs").WithWhenHelper()

	template := `
	{{#when user_name "eq" "ADMIN"}}
	<p>Hello, {{user_name}}</p>
	{{else}}
	<p>Hello,</p>
	{{/when}}

	{{#when pract_name "noteq" ""}}
	<p>Hello, {{pract_name}}</p>
	{{else}}
	<p>Hello,</p>
	{{/when}}

	{{#when count1 "gt" count2}}
	<p>Admin has got, ${{count1}}</p>
	{{else}}
	<p>Admin has got, ${{count2}}</p>
	{{/when}}

	{{#when count1 "lt" count2}}
	<p>You have got, ${{count1}}</p>
	{{else}}
	<p>You have got, ${{count2}}</p>
	{{/when}}
`

	params := map[string]interface{}{
		"user_name":  "ADMIN",
		"pract_name": "Doe",
		"count1":     "100",
		"count2":     "200",
	}

	content := html.UnescapeString(template)
	content = engine.RenderDynamic(content, params)

	err := engine.Render(&buf, "email", map[string]interface{}{"Content": content})
	assert.NoError(t, err)

	expect := `<p>Hello, ADMIN</p><p>Hello, Doe</p><p>Admin has got, $200</p><p>You have got, $100</p>`
	result := trim(buf.String())

	assert.Equal(t, expect, result)
}
