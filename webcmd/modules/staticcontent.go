package modules

import (
    "../staticcontent"
    "io/ioutil"
    "net/http"
    "html/template"
)

// This struct implements modules.Module
type StaticContentModule struct {
    server *staticcontent.Server
}

func NewStaticContentModule(server *staticcontent.Server) *StaticContentModule {
    return &StaticContentModule{
        server: server,
    }
}

func (m *StaticContentModule) Init() error {
    return nil
}

var STATIC_CONTENT_TEMPLATE_FILE = "templates/static_content.html.template"

func (m *StaticContentModule) Name() string {
    return "Static Content Module"
}

func (m *StaticContentModule) Commands() []string {
    return []string{"static", "files"}
}

func (m *StaticContentModule) RunCommand(command string, args string) (template.HTML, error) {
    return m.List()
}

func (m *StaticContentModule) RunEvent(req *http.Request) (template.HTML, error) {
    return m.List()
}

func (m *StaticContentModule) List() (template.HTML, error) {
    template_content, err := ioutil.ReadFile(STATIC_CONTENT_TEMPLATE_FILE)
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
    return w.S, nil
}