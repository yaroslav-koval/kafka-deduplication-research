package handlebars

import (
	"context"
	"fmt"
	"io"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/template/utils"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/aymerick/raymond"
)

// Engine struct
type Engine struct {
	ctx context.Context
	// views folder
	directory string
	// http.FileSystem supports embedded files
	fileSystem http.FileSystem
	// views extension
	extension string
	// layout variable name that encapsulates the template
	layout string
	// determines if the engine parsed all templates
	loaded bool
	// reload on each render
	reload bool
	// debug prints the parsed templates
	debug bool
	// lock for funcmap and templates
	mutex sync.RWMutex
	// template funcmap
	funcmap map[string]interface{}
	// templates
	Templates map[string]*raymond.Template
}

// New returns a Handlebar render engine for Fiber
func New(directory, extension string) *Engine {
	engine := &Engine{
		ctx:       context.Background(),
		directory: directory,
		extension: extension,
		layout:    "embed",
		funcmap:   make(map[string]interface{}),
	}

	return engine
}

func NewFileSystem(fs http.FileSystem, extension string) *Engine {
	engine := &Engine{
		directory:  "/",
		fileSystem: fs,
		extension:  extension,
		layout:     "embed",
		funcmap:    make(map[string]interface{}),
	}

	return engine
}

// WithContext used custom context
func (e *Engine) WithContext(ctx context.Context) {
	e.ctx = ctx
}

func (e *Engine) WithWhenHelper() *Engine {
	engine := e.AddFunc("when", CustomIfHelper)

	raymond.RegisterHelpers(engine.funcmap)

	return engine
}

// Delims sets the action delimiters to the specified strings, to be used in
// templates. An empty delimiter stands for the
// corresponding default: {{ or }}.
func (e *Engine) Delims(left, right string) *Engine {
	fmt.Println("delims: this method is not supported for handlebars")
	return e
}

// Layout defines the variable name that will encapsulate the template
func (e *Engine) Layout(key string) *Engine {
	e.layout = key
	return e
}

// AddFunc adds the function to the template's function map.
func (e *Engine) AddFunc(name string, fn interface{}) *Engine {
	e.mutex.Lock()
	e.funcmap[name] = fn
	e.mutex.Unlock()

	return e
}

// Reload if set to true the templates are reloading on each render,
// use it when you're in development, and you don't want to restart
// the application when you edit a template file.
func (e *Engine) Reload(enabled bool) *Engine {
	e.reload = enabled
	return e
}

// Debug will print the parsed templates when Load is triggered.
func (e *Engine) Debug(enabled bool) *Engine {
	e.debug = enabled
	return e
}

// Load parses the templates to the engine.
func (e *Engine) Load() error {
	var err error
	// race safe
	e.mutex.Lock()
	defer e.mutex.Unlock()

	// Set template settings
	e.Templates = make(map[string]*raymond.Template)

	// Loop trough each directory and register template files
	walkFn := func(path string, info os.FileInfo, err error) error {
		// Return error if exist
		if err != nil {
			return err
		}
		// Skip file if it's a directory or has no file info
		if info == nil || info.IsDir() {
			return nil
		}
		// Skip file if it does not equal the given template extension
		if len(e.extension) >= len(path) || path[len(path)-len(e.extension):] != e.extension {
			return nil
		}
		// Get the relative file path
		// ./views/html/index.tmpl -> index.tmpl
		rel, err := filepath.Rel(e.directory, path)
		if err != nil {
			return err
		}
		// Reverse slashes '\' -> '/' and
		// partials\footer.tmpl -> partials/footer.tmpl
		name := filepath.ToSlash(rel)
		// Remove ext from name 'index.tmpl' -> 'index'
		name = strings.TrimSuffix(name, e.extension)

		// Read the file
		buf, err := utils.ReadFile(path, e.fileSystem)
		if err != nil {
			return err
		}
		// Create new template associated with the current one
		// This enable use to invoke other templates {{ template .. }}
		tmpl, err := raymond.Parse(string(buf))
		if err != nil {
			return err
		}

		e.Templates[name] = tmpl

		// Debugging
		if e.debug {
			fmt.Printf("views: parsed template: %s\n", name)
		}

		return err
	}

	if e.fileSystem != nil {
		err = utils.Walk(e.fileSystem, e.directory, walkFn)
	} else {
		err = filepath.Walk(e.directory, walkFn)
	}

	// Link templates with each-other
	for i := range e.Templates {
		for n, t := range e.Templates {
			e.Templates[i].RegisterPartialTemplate(n, t)
		}
	}

	e.loaded = true

	if err != nil {
		return e.wrapError("handlebars Load", err)
	}

	return nil
}

// RenderDynamic rendering dynamic html used custom params
func (e *Engine) RenderDynamic(source string, params map[string]interface{}) string {
	tpl := raymond.MustParse(source)

	return tpl.MustExec(params)
}

// Render executes render the template by name
//
//nolint:nestif
func (e *Engine) Render(out io.Writer, template string, binding interface{}, layout ...string) error {
	var err error

	if !e.loaded || e.reload {
		if e.reload {
			e.loaded = false
		}

		err = e.Load()
		if err != nil {
			return err
		}
	}

	tmpl := e.Templates[template]
	if tmpl == nil {
		return cerror.NewF(e.ctx, cerror.KindInternal, "render: template %s does not exist", template).
			LogError()
	}

	parsed, err := tmpl.Exec(binding)
	if err != nil {
		return cerror.NewF(e.ctx, cerror.KindInternal, "render: %v", err).
			LogError()
	}

	if len(layout) > 0 && layout[0] != "" {
		lay := e.Templates[layout[0]]
		if lay == nil {
			return cerror.NewF(e.ctx, cerror.KindInternal, "render: layout %s does not exist", layout[0]).
				LogError()
		}

		var bind map[string]interface{}
		if m, ok := binding.(map[string]interface{}); ok {
			bind = m
		} else {
			bind = make(map[string]interface{}, 1)
		}

		bind[e.layout] = raymond.SafeString(parsed)

		parsed, err = lay.Exec(bind)
		if err != nil {
			return cerror.NewF(e.ctx, cerror.KindInternal, "render: %v", err).
				LogError()
		}

		if _, err = out.Write([]byte(parsed)); err != nil {
			return cerror.NewF(e.ctx, cerror.KindInternal, "render: %v", err).
				LogError()
		}

		return nil
	}

	if _, err = out.Write([]byte(parsed)); err != nil {
		return cerror.NewF(e.ctx, cerror.KindInternal, "render: %v", err).
			LogError()
	}

	return err
}

func (e *Engine) wrapError(m string, err error) error {
	switch vErr := err.(type) {
	case *cerror.CError:
		return vErr
	default:
		return cerror.NewF(e.ctx, cerror.KindInternal, "%s: %s", m, vErr).LogError()
	}
}
