package assets

import (
	_ "embed"
	"fmt"
	"io"
	"text/template"
)

//go:embed command.tmpl
var commandTemplateContents string

var commandTemplate = template.Must(template.New("").Parse(commandTemplateContents))

type ScriptArgs struct {
	PackagePath   string
	CommandName   string
	IsolateModule bool
}

func WriteScriptContents(w io.Writer, args ScriptArgs) error {
	if err := commandTemplate.Execute(w, args); err != nil {
		return fmt.Errorf("could not populate script contents: %w", err)
	}
	return nil
}
