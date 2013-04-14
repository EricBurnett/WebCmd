package modules

import (
    "flag"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "path/filepath"
    "html/template"
    "time"
)

var gs_path = flag.String("gs_path",
    "C:\\Users\\Somebody\\AppData\\Roaming\\GroovesharkDesktop." +
    "LOTSOFHEXDIGITS.1\\Local Store",
	"Path to GrooveShark control directory.")
var gs_control_file = flag.String("gs_control_file", "shortcutAction.txt",
	"GrooveShark control file.")

type Page struct {
	Message string
}

// This struct implements modules.Module
type GSModule struct {
    // Channel to use for tracking message requests.
    MessageChannel chan string
    
    // Path to GrooveShark file to write instructions to.
    File string
}

func NewGSModule() *GSModule {
    return &GSModule{
        MessageChannel: make(chan string, 100),
    }
}

func (m *GSModule) Init() error {
    m.File = filepath.Join(*gs_path, *gs_control_file)

	// Initialize GrooveShark.
    _, err := os.Stat(m.File)
	if err != nil {
        return err
    }
	go m.pushMessages()
    return nil
}

var GS_TEMPLATE_FILE = "templates/gs.html.template"

func (m *GSModule) Name() string {
    return "GrooveShark Controller"
}

func (m *GSModule) Commands() []string {
    return []string{"gs", "grooveshark", "music"}
}

func (m *GSModule) RunCommand(command string, args string) (template.HTML, error) {
    return m.ComposeForm("")
}

func (m *GSModule) RunEvent(req *http.Request) (template.HTML, error) {
	choice := req.FormValue("gs_choice")
	switch choice {
	case "Previous":
		m.MessageChannel <- "previoussong"
	case "Next":
		m.MessageChannel <- "next"
	case "Play/Pause":
		m.MessageChannel <- "playpause"
	case "Volume up":
		m.MessageChannel <- "volumeup"
	case "Volume down":
		m.MessageChannel <- "volumedown"
	}
    
    return m.ComposeForm(choice)
}

func (m *GSModule) ComposeForm(message string) (template.HTML, error) {
    template_content, err := ioutil.ReadFile(GS_TEMPLATE_FILE)
    if err != nil {
        return "", err
    }
    var gsTemplate = template.New("GS template")
    gsTemplate, err = gsTemplate.Parse(string(template_content))
    if err != nil {
        return "", err
    }
	p := &Page{Message: message}
    var w HTMLWriter
    gsTemplate.Execute(&w, p)
    return w.S, nil
}

// Write a single message to GrooveShark.
func (m *GSModule) writeMessageToGS(message string) bool {
	log.Print("Writing message: ", message)
	var last_err error = nil
	for i := 0; i < 50; i++ { // Retry a write up to 50 times (5s)
		file, err := os.OpenFile(m.File, os.O_WRONLY|os.O_APPEND, 0)
		if err != nil {
			last_err = err
			time.Sleep(100*time.Millisecond)
			continue
		}

		_, err = fmt.Fprintln(file, message)
		file.Close()
		if err != nil {
			last_err = err
			time.Sleep(100*time.Millisecond)
			continue
		}
		log.Print("Write succeeded")
		return true
	}
	log.Print("Could not write message to GrooveShark: ", last_err)
	return false
}

func (m *GSModule) pushMessages() {
	for message := range m.MessageChannel {
		// Take all pending messages as well
		var lookForMore = true
		for lookForMore {
			select {
			case next := <-m.MessageChannel:
				message += "\n" + next
			default:
				lookForMore = false
			}
		}
		// Write all messages at once
		m.writeMessageToGS(message)
	}
}