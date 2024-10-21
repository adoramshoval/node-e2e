package template

import (
	"bytes"
	"io"
	"text/template"
)

const (
	tmplName string = "Template"
)

func Template(values map[string]interface{}, tmplData []byte) (io.Reader, error) {
	// Parse the template
	tmpl, err := template.New(tmplName).Parse(string(tmplData))
	if err != nil {
		return nil, err
	}

	// Buffer to hold the processed YAML after applying the template
	var processed bytes.Buffer

	// Execute the template with the map of values
	if err := tmpl.Execute(&processed, values); err != nil {
		return nil, err
	}

	// Return the YAML bytes
	return bytes.NewReader(processed.Bytes()), nil
}
