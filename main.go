package main

import (
	"flag"
	"fmt"
	"github.com/EricBurnett/WebCmd/platform"
	"log"
	"net"
	"os"
	"path/filepath"
)

var host = flag.String("host", ":8080", "Server address to host on")
var window = flag.Bool("window", true,
	"Try to start a GUI window. May not be available on all platforms.")

func main() {
	flag.Parse()

	// Redirect log output to file.
	_, prog := filepath.Split(os.Args[0])
	log_path := filepath.Join(os.TempDir(), prog+".INFO")
	w, err := os.OpenFile(log_path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err == nil {
		log.Print("Setting log output to ", log_path)
		log.SetOutput(w)
		log.Print("Logfile created")
	} else {
		log.Print("Failed to redirect output: ", err)
	}

	serverURL := ServeAsync()

	block := !*window
	if *window {
		if err := platform.WebviewWindow(serverURL); err != nil {
			log.Print(err)
			block = true
		}
	}
	// If we're not starting a window (or it failed), block forever instead.
	if block {
		select {}
	}
	log.Println("Main done, terminating.")
}

// Returns the address the server is running on, as a host:port string.
func ServeAsync() string {
	s := CreateServer(*host)
	go func() {
		err := s.ListenAndServe()
		if err != nil {
			log.Fatal("Could not listen on address ", *host, ", ", err)
		}
	}()
	addr, err := net.ResolveTCPAddr("tcp4", s.Addr)
	if err != nil {
		log.Fatal("Could not resolve address of local server")
	}
	if addr.IP == nil {
		addr.IP = net.IP{127, 0, 0, 1}
	}
	return fmt.Sprintf("http://%v:%v", addr.IP, addr.Port)
}
