package modules

import (
    "../staticcontent"
    "html/template"
    "log"
    "net/http"
)

// A module corresponds to a single web command. Modules can be triggered by
// strings (commands), as well as by get/post events (form events). These
// correspond to (web) command-line invocations, and then interaction with
// the module after is has been invoked.
type Module interface {    
    // Initializes the module.
    Init() error
    
    // The name of this module.
    Name() string
    
    // The preferred command strings for the module. Must be single words, and
    // there may be multiple. e.g. ["gs", "grooveshark", "music"]
    Commands() []string

    // Runs a command string. The triggering command is passed as one parameter,
    // The arguments (if any) as the second. Returns HTML to be inserted into
    // the main page template, or an error.
    RunCommand(command string, args string) (template.HTML, error)
    
    // Runs a command event (get or post). Returns HTML to be inserted into
    // the main page template, or an error.
    RunEvent(*http.Request) (template.HTML, error)
}

type HTMLWriter struct {
    S template.HTML
}

func (m *HTMLWriter) Write(p []byte) (n int, err error) {
    m.S = m.S + template.HTML(p)
    return len(p), nil
}

func InstalledModules(ss *staticcontent.Server) []Module {
    m := []Module{}
    tryAdd(&m, NewGSModule())
    tryAdd(&m, NewStaticContentModule(ss))
    return m
}

func tryAdd(m *[]Module, module Module) {
    err := module.Init()
    if err != nil {
        log.Println(err)
    } else {
        *m = append(*m, module)
    }
}