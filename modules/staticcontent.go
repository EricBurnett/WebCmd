package modules

import (
	"github.com/EricBurnett/WebCmd/resources"
	"github.com/EricBurnett/WebCmd/staticcontent"
	"html/template"
	"net/http"
)

// StaticContentModule implements modules.Module and provides a listing of all
// static content roots currently mapped.
type StaticContentModule struct {
	server *staticcontent.Server
}

// Returns a new StaticContentModule, showing content for the provided static
// content server.
func NewStaticContentModule(server *staticcontent.Server) *StaticContentModule {
	return &StaticContentModule{
		server: server,
	}
}

// Initializes the StaticContentModule.
func (m *StaticContentModule) Init() error {
	return nil
}

// The name of this module.
func (m *StaticContentModule) Name() string {
	return "Static Content Module"
}

// The command hooks to install under.
func (m *StaticContentModule) Commands() []string {
	return []string{"static", "files"}
}

// RunCommand runs a single command. This module always just prints a listing
// of mapped paths.
func (m *StaticContentModule) RunCommand(command string, args string) (template.HTML, error) {
	return m.List()
}

// RunEvent responds to module events.  This module always just prints a listing
// of mapped paths.
func (m *StaticContentModule) RunEvent(req *http.Request) (template.HTML, error) {
	return m.List()
}

var STATIC_CONTENT_TEMPLATE_FILE = "templates/static_content.html.template"

// Produces a listing in HTML.
func (m *StaticContentModule) List() (template.HTML, error) {
	template_content, err := resources.Load(STATIC_CONTENT_TEMPLATE_FILE)
	if err != nil {
		return "", err
	}
	var staticContentTemplate = template.New("Static Content template")
	staticContentTemplate, err = staticContentTemplate.Parse(string(template_content))
	if err != nil {
		return "", err
	}

	var w HTMLWriter
	staticContentTemplate.Execute(&w, m.server.Roots())
	return w.HTML(), nil
}
