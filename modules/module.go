package modules

import (
	"../staticcontent"
	"html/template"
	"log"
	"net/http"
)

// A module corresponds to a single web command or tool. Modules can be
// triggered by strings (commands), as well as by get/post events (form events).
// These correspond to (web) command-line invocations, and interaction with
// the module directly. Some modules may only do meaningful work via one of
// these interfaces.
type Module interface {
	// Initializes the module. If an error is returned, the module will not be
    // installed.
	Init() error

	// The name of this module. May be called before Init.
	Name() string

	// The preferred command strings for the module. May be called before Init.
    // Output be single words, and there may be multiple. e.g.
    // ["gs", "grooveshark", "music"]
	Commands() []string

	// Runs a command string. The triggering command is passed as one parameter,
	// The arguments (if any) as the second. Returns HTML to be inserted into
	// the main page template, or an error.
	RunCommand(command string, args string) (template.HTML, error)

	// Runs a command event (get or post). Returns HTML to be inserted into
	// the main page template, or an error.
	RunEvent(*http.Request) (template.HTML, error)
}

// Returns instances of all the modules available. Returned modules are
// initialized already, and any that failed to init have been filtered out.
func InstalledModules(ss *staticcontent.Server) []Module {
	m := []Module{}
	tryAdd(&m, NewGSModule())
	tryAdd(&m, NewStaticContentModule(ss))
	return m
}

// Tries to add a module to the list, calling Init first. if Init fails the
// module is not added.
func tryAdd(m *[]Module, module Module) {
	err := module.Init()
	if err != nil {
		log.Println(err)
	} else {
		*m = append(*m, module)
	}
}

// Helper that implements io.Writer and interprets the results as HTML. For
// executing templates and returning the result.
type HTMLWriter struct {
	h template.HTML
}

// Adds HTML to the writer.
func (m *HTMLWriter) Write(p []byte) (n int, err error) {
    // TODO: use something more efficient.
	m.h = m.h + template.HTML(p)
	return len(p), nil
}

// Returns all the HTML added since creation.
func (m *HTMLWriter) HTML() template.HTML {
    return m.h
}
