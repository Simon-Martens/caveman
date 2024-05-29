// The source for this file is pocketbase/tools/template
package templates

import (
	"bytes"
	"errors"
	"html/template"
)

// Renderer defines a single parsed template.
// It's a wrapper around html/templates
type Template struct {
	Directory  *Dir
	Template   *template.Template
	Headers    *template.Template
	parseError error
}

// Render executes the template with the specified data as the dot object
// and returns the result as plain string.
func (r *Template) Render(data any) (string, error) {
	if r.parseError != nil {
		return "", r.parseError
	}

	if r.Template == nil {
		return "", errors.New("invalid or nil template")
	}

	buf := new(bytes.Buffer)

	if err := r.Template.Execute(buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (r *Template) RenderPartial(name string, data any) (string, error) {
	if r.parseError != nil {
		return "", r.parseError
	}

	if r.Template == nil {
		return "", errors.New("invalid or nil template")
	}

	buf := new(bytes.Buffer)

	if err := r.Template.ExecuteTemplate(buf, name+"."+FILE_FORMAT, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
