// +build windows

package platform

import (
	"errors"
	"log"
)

// https://github.com/lxn/walk
import (
	"github.com/lxn/walk"
	"github.com/lxn/walk/declarative"
)

// Spawns a window containing a webview pointed at the given URL, and waits for
// it to be closed. If the webview cannot be opened, returns an error
// immediately.
func WebviewWindow(serverURL string) (e error) {
	var mainWindow *walk.MainWindow
	var webView *walk.WebView

	log.Print("Starting to create webview window")
	if err := (declarative.MainWindow{
		AssignTo: &mainWindow,
		Title:    "WebCmd Webview",
		MinSize:  declarative.Size{600, 400},
		Size:     declarative.Size{800, 600},
		Visible:  true,
		Layout:   declarative.HBox{},
		Children: []declarative.Widget{
			declarative.WebView{
				AssignTo: &webView,
			},
		},
	}.Create()); err != nil {
		log.Println("Error while creating window. Usually happens when " +
			"compiled without -ldflags=\"-Hwindowsgui\" or a manifest. " +
			"see make.bat to compile with window support on windows.")
		return errors.New("Unable to create webview window")
	}
	log.Print("Create complete, initializing webView with URL ", serverURL)
	webView.SetURL(serverURL)

	mainWindow.Run()
	log.Print("Webview closed.")
	return nil
}
