package main

import (
    "./modules"
    "./staticcontent"
    "fmt"
    "html/template"
    "io/ioutil"
    "log"
    "net/http"
    "strings"
)

type WebCmdServer struct {
    http.Server
    errorTemplate *template.Template
    modules map[string] modules.Module
    staticContentServer *staticcontent.Server
}

// Returns a server, loaded and ready to go.
func CreateServer(host string) *WebCmdServer {
    var err error
    
    var errorTemplate = template.New("Error template")
    errorTemplate, err = errorTemplate.Parse(ERROR_TEMPLATE_STR)
    if err != nil {
        log.Fatal("Error creating error template: ", err)
    }
    
    server := WebCmdServer{
        Server: http.Server{
            Addr:       host,
        },
        errorTemplate:      errorTemplate,
        modules: make(map[string] modules.Module),
    }
    
    server.staticContentServer = staticcontent.NewServer("/static_root", server.Server)
    if err = staticcontent.AddCsvPaths(server.staticContentServer); err != nil {
        log.Println("Error installing paths from csv:", err)
    }
    allModules := modules.InstalledModules(server.staticContentServer)
    
    for _, module := range allModules {
        for _, command := range module.Commands() {
            if _, has := server.modules[command]; has {
                log.Println("Handler already present for", command,
                            "; not installing", module.Name())
                continue 
            }
            path := fmt.Sprintf("/%v", command)
            http.Handle(path, http.HandlerFunc(server.bareModuleHandler(command, module)))
            log.Println("Installing", module.Name(), "at path", path)
            server.modules[command] = module
        }
    }

	http.Handle("/", http.HandlerFunc(server.serveRoot()))
    return &server
}

type Page struct {
    Title string        // Page title
	Message string      // Text message to be printed at the top
    Body template.HTML  // HTML body to be injected into the middle
    Path string         // Path the main form should redirect to (skip first /)
    QueryString string  // The query string to add to the form
    Command string      // The command to delegate to for form executions
}

var BARE_MODULE_FILE = "templates/bare_module.html.template"

func (server WebCmdServer) bareModuleHandler(command string, m modules.Module) func(http.ResponseWriter, *http.Request) {
    return func(w http.ResponseWriter, req *http.Request) {
        logRequest(req)
        template_content, err := ioutil.ReadFile(BARE_MODULE_FILE)
        if err != nil {
            server.PrintError(w, err)
            return
        }
        var bareModuleTemplate = template.New("Bare module template")
        bareModuleTemplate, err = bareModuleTemplate.Parse(string(template_content))
        if err != nil {
            server.PrintError(w, err)
            return
        }
        body, err := m.RunEvent(req)
        if err != nil {
            server.PrintError(w, err)
            return
        }
        query := req.FormValue("q")
        page := Page{
            Title: m.Name(), Body: body, Path: command,
            Command: command, QueryString: query}
        bareModuleTemplate.Execute(w, &page)
    }
}

var ROOT_TEMPLATE_FILE = "templates/root.html.template"

func (server WebCmdServer) serveRoot() func(http.ResponseWriter, *http.Request) {
    return func(w http.ResponseWriter, req *http.Request) {
        logRequest(req)

        source := req.FormValue("source")
        query := strings.TrimSpace(req.FormValue("q"))
        var message string
        var command string
        var body template.HTML
        var err error
        
        if source == "" || (source == "query" && (query == "" || strings.ToLower(query) == "help")) {
            body, err = ModuleList(server.modules)
        } else if source == "query" {
            queryPieces := strings.SplitN(query, " ", 2)
            if len(queryPieces) == 1 {
                queryPieces = append(queryPieces, "")
            }
            if module, has := server.modules[queryPieces[0]]; has {
                command = queryPieces[0]
                body, err = module.RunCommand(queryPieces[0], queryPieces[1])
            } else {
                message = "Module not found for query. Try again?"
            }
        } else {
            if module, has := server.modules[source]; has {
                command = source
                body, err = module.RunEvent(req)
            } else {
                message = "Requested module not found. Try a query instead!"
            }
        }
        
        if err != nil {
            log.Println(err)
            message = err.Error()
        }
    
        template_content, err := ioutil.ReadFile(ROOT_TEMPLATE_FILE)
        if err != nil {
            server.PrintError(w, err)
            return
        }
        var rootTemplate = template.New("Root template")
        rootTemplate, err = rootTemplate.Parse(string(template_content))
        if err != nil {
            server.PrintError(w, err)
            return
        }
        page := Page{
            Title: "root", QueryString: query, Message: message, Body: body,
            Command: command}
        rootTemplate.Execute(w, &page)
    }
}

type PageError struct {
    Error string
}
var ERROR_TEMPLATE_STR = "<html><head><title>WebCmd - Error</title></head>" +
    "<body>{{printf \"%s\" .Error |html}}</body></html>"

func (server WebCmdServer) PrintError(w http.ResponseWriter, e error) {
    log.Println(e)
    server.errorTemplate.Execute(w, &PageError{Error: e.Error()})
}

// Write the information from a request to the output file.
func logRequest(req *http.Request) {
	var form = ""
	if req.Method == "POST" {
		req.ParseForm()
		form = " Form: " + fmt.Sprintf("%v", req.Form)
	}
	log.Printf("Request: %v %v %v %v%v %v", req.Method, req.Host, req.URL,
		req.Proto, form, "From "+req.RemoteAddr)
}