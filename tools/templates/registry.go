// Package template is a thin wrapper around the standard html/template
// and text/template packages that implements a convenient registry to
// load and cache templates on the fly concurrently.
//
// The source for this file is located in pocketbase/tools/template
//
// Example:
//
//	registry := template.NewRegistry()
//
//	html1, err := registry.LoadFiles(
//		// the files set wil be parsed only once and then cached
//		"layout.html",
//		"content.html",
//	).Render(map[string]any{"name": "John"})
//
//	html2, err := registry.LoadFiles(
//		// reuse the already parsed and cached files set
//		"layout.html",
//		"content.html",
//	).Render(map[string]any{"name": "Jane"})
package templates

import (
	"errors"
	"html/template"
	"io/fs"
	"log"
	"path/filepath"
	"strings"

	"github.com/Masterminds/sprig/v3"
	"github.com/Simon-Martens/caveman/tools/store"
)

const (
	FILE_FORMAT        = "tmpl"
	GLOBAL_FILE_PREFIX = "_"
	END_BLOCK          = "{{end}}"
	DEFAULT_ROOT       = `<!DOCTYPE html>
<html>

<head>
{{block "head" .}}
<!-- Default Head elements -->
{{end}}
</head>

<body>
    {{block "body" .}}
    <!-- Default app body... -->
    {{end}}
</body>

</html>`
)

type RegistryOptions struct {
	Extension     string
	InhFilePrefix string
}

func DefaultRegistryOptions() RegistryOptions {
	return RegistryOptions{
		Extension:     FILE_FORMAT,
		InhFilePrefix: GLOBAL_FILE_PREFIX,
	}
}

// Registry defines a templates registry that is safe to be used by multiple goroutines.
//
// Use the Registry.Load* methods to load templates into the registry.
type Registry struct {
	options RegistryOptions

	routes fs.FS
	cache  *store.Store[*Template]
	funcs  template.FuncMap
}

type Dir struct {
	root       string
	body       string
	head       string
	headers    string
	paths      []string
	components []string
	basepath   string
}

func (dir Dir) String() string {
	return "Template\nRoot=" + dir.root + "\nBody=" + dir.body + "\nHead=" + dir.head + "\nHeaders=" + dir.headers + "\nBasepath=" + dir.basepath
}

// NewRegistry creates and initializes a new templates registry with
// some defaults (eg. global "raw" template function for unescaped HTML).
//
// Use the Registry.Load* methods to load templates into the registry.
func NewRegistry(routes fs.FS, options RegistryOptions) *Registry {
	return &Registry{
		options: options,
		routes:  routes,
		cache:   store.New[*Template](nil),
		funcs: template.FuncMap{
			"raw": func(str string) template.HTML {
				return template.HTML(str)
			},
		},
	}
}

func (r *Registry) read_dir(path string) (*Dir, error) {
	ext := r.options.Extension
	pref := r.options.InhFilePrefix
	dir := Dir{
		basepath: path,
	}

	glob := path
	if len(path) > 0 {
		glob = path + "/*." + ext
	} else {
		glob = path + "*." + ext
	}

	log.Println("Looking for files: " + glob)

	matches, err := fs.Glob(r.routes, glob)
	for _, m := range matches {
		f, err := fs.Stat(r.routes, m)
		if err != nil || f.IsDir() {
			log.Println("Could not stat " + m)
			continue
		}

		log.Println("Match: " + m)
		name := f.Name()

		if name == "body."+ext {
			dir.body = m
		} else if name == "head."+ext {
			dir.head = m
		} else if name == "headers."+ext {
			dir.headers = m
		} else if name == pref+"root."+ext {
			// TODO Root must not be appended unconditionally. It must have a head and a body
			// TODO we search the tree for body and head template invocations
			str, err := r.file_to_string(m)
			if err != nil {
				log.Panicln(err.Error())
			}

			tmpl, err := filename_string_to_template(m, str)
			if err != nil {
				log.Panicln(err.Error())
			}

			body := tmpl.Lookup("body")
			head := tmpl.Lookup("head")

			if head == nil || body == nil {
				log.Panicln("The Root Template " + m + " must define a body and head template using the {{block}} syntax.")
			}

			dir.root = m
		} else if strings.HasPrefix(name, r.options.InhFilePrefix) {
			log.Println("Adding component " + m)
			dir.components = append(dir.components, m)
		} else {
			dir.paths = append(dir.paths, m)
		}

	}
	if err != nil {
		return nil, errors.New("Keine Template-Files im Verzeichnis " + path + " gefunden.")
	}

	return &dir, nil
}

func (r *Registry) read_dir_recursively(path string) (*Dir, error) {
	dir, err := r.read_dir(path)
	if err != nil {
		return nil, err
	}

	d, _ := filepath.Split(path)
	for len(dir.root) == 0 || len(d) > 0 {
		d = filepath.Clean(d)
		parent, err := r.read_dir(d)
		if err != nil {
			return nil, err
		}

		if len(parent.root) > 0 {
			dir.root = parent.root
		}

		if len(parent.components) > 0 {
			dir.components = append(dir.components, parent.components...)
		}

		d, _ = filepath.Split(d)
	}

	return dir, nil
}

// AddFuncs registers new global template functions.
//
// The key of each map entry is the function name that will be used in the templates.
// If a function with the map entry name already exists it will be replaced with the new one.
//
// The value of each map entry is a function that must have either a
// single return value, or two return values of which the second has type error.
//
// Example:
//
//	r.AddFuncs(map[string]any{
//	  "toUpper": func(str string) string {
//	      return strings.ToUppser(str)
//	  },
//	  ...
//	})
func (r *Registry) AddFuncs(funcs map[string]any) *Registry {
	for name, f := range funcs {
		r.funcs[name] = f
	}

	return r
}

// This function parses a direcory structure for dirs that contain tmpl files.
// We need to kindof create a dependency graph bc the sequence of documents passed
// into the templating engine is significant.
// We can have layouts or nested layouts in a hierachical structure, which get passed to
// subdirectories, if they don't contain any layout files themselves.
// Parse sequence is significant. We parse:
//  1. layout.tmpl
//  2. body.tmpl
//  3. head.tmpl
//  4. all other .tmpl files
//
// header.tmpl files are parsed seperately into response headers
// (we skip the dir if no index + no body + no header is present)
func (r *Registry) LoadDir(path string) (*Template, error) {
	ext := r.options.Extension

	path = normalize_path(path)
	path = filepath.Clean(path)
	found := r.cache.Get(path)
	if found != nil {
		log.Println("Found template in cache for " + path)
		return found, nil
	}

	dir, err := r.read_dir_recursively(path)
	if err != nil {
		return nil, err
	}

	tpl := template.New(r.options.InhFilePrefix + "root." + ext).Funcs(r.funcs)
	if len(dir.root) == 0 {
		tpl, err = tpl.Parse(DEFAULT_ROOT)
		if err != nil {
			return nil, err
		}
	} else {
		tpl, err = tpl.ParseFS(r.routes, dir.root)
		if err != nil {
			return nil, err
		}
	}

	for _, file := range dir.components {
		tpl, err = tpl.ParseFS(r.routes, file)
		if err != nil {
			return nil, err
		}
	}

	if len(dir.head) > 0 {
		str, err := r.find_create_def(dir.head, "head")
		if err != nil {
			// TODO: Panic here?
			return nil, err
		}
		tpl, err = tpl.Parse(*str)
		if err != nil {
			// TODO: Panic here?
			return nil, err
		}
	}

	if len(dir.body) > 0 {
		str, err := r.find_create_def(dir.body, "body")
		if err != nil {
			// TODO: Panic here?
			return nil, err
		}
		tpl, err = tpl.Parse(*str)
		if err != nil {
			// TODO: Panic here?
			return nil, err
		}
	}

	for _, file := range dir.paths {
		tpl, err = tpl.ParseFS(r.routes, file)
		if err != nil {
			return nil, err
		}
	}

	temp := Template{
		Directory: dir,
		Template:  tpl,
	}

	if len(dir.headers) > 0 {
		temp.Headers, err = template.New("headers."+ext).Funcs(r.funcs).Funcs(sprig.FuncMap()).ParseFS(r.routes, dir.headers)
		if err != nil {
			return nil, err
		}
	}

	r.cache.Set(path, &temp)

	return &temp, nil
}

func (r *Registry) LoadFile(path string) (*Template, error) {
	ext := r.options.Extension
	path = normalize_path(path)
	path = filepath.Clean(path)
	found := r.cache.Get(path)
	if found != nil {
		log.Println("Found template in cache for " + path)
		return found, nil
	}

	d, f := filepath.Split(path)
	f = f + "." + ext
	dir, err := r.read_dir_recursively(filepath.Clean(d))
	if err != nil {
		return nil, err
	}

	tpl := template.New(f).Funcs(r.funcs)

	tpl, err = tpl.ParseFS(r.routes, path+"."+ext)
	if err != nil {
		return nil, err
	}

	for _, file := range dir.components {
		tpl, err = tpl.ParseFS(r.routes, file)
		if err != nil {
			return nil, err
		}
	}

	temp := Template{
		Directory: dir,
		Template:  tpl,
	}

	r.cache.Set(path, &temp)
	return &temp, nil
}

func (r *Registry) find_create_def(filename string, tofind string) (*string, error) {
	str, err := r.file_to_string(filename)
	if err != nil {
		return nil, err
	}

	tmp, err := filename_string_to_template(filename, str)
	if err != nil {
		return nil, err
	}

	found := tmp.Lookup(tofind)

	if found != nil {
		return str, nil
	}

	ret := create_define_start_block(tofind) + *str + END_BLOCK
	return &ret, nil
}

func create_define_start_block(todefine string) string {
	return `{{define "` + todefine + `"}}`
}

func (r *Registry) file_to_string(filename string) (*string, error) {
	fc, err := fs.ReadFile(r.routes, filename)
	if err != nil {
		errors.New("File " + filename + " can't be opened for reading. " + err.Error())
	}
	fcs := string(fc[:])
	return &fcs, nil
}

func filename_string_to_template(filename string, input *string) (*template.Template, error) {
	name := filepath.Base(filename)
	tmpl := template.New(name)
	tmpl, err := tmpl.Parse(*input)
	if err != nil {
		return nil, errors.New("File " + filename + " is not a valid Template. " + err.Error())
	}
	return tmpl, nil
}

func normalize_path(str string) string {
	// Path that starts with a slash are causing wild behaviour
	return strings.TrimLeft(filepath.ToSlash(filepath.Clean(str)), "/")
}
