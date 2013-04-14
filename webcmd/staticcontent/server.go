package staticcontent

import (
    "../platform"
    "errors"
    "flag"
    "fmt"
    "html/template"
    "io/ioutil"
    "log"
    "net/http"
    "os/exec"
    "path"
    "path/filepath"
    "sort"
    "strings"
    "time"
)

var custom_video_player = flag.Bool("custom_video_player", true,
	"Embed videos in a custom video player.")

var transcoder = flag.String("transcoder", "ffmpeg",
    "The transcoder to use, either as a fully qualified path or as an " +
    "executable on the path.")
var transcode_settings = flag.String("transcode_settings",
    "-vcodec libvpx -threads 0 -bufsize 100m -b:v 3000k -bt 300k -acodec " +
    "libvorbis -ab 96k -ac 2 -f webm -quality realtime -",
    "Transcode settings to pass to the transcoder. Note that the transcoder " +
    "must be configured to write the result to stdout.")
var transcode_content_type = flag.String("transcode_content_type", "webm",
    "The Content-Type used for transcoded video output.")
var verbose_transcode_output = flag.Bool("verbose_transcode_output", false,
    "Log verbose transcode output.")

type Server struct {
    prefix string
    httpServer http.Server
    installedPaths map[string] string
}

// Creates a new Server. On request, this object will install new
// file system handlers under prefix. E.g. if prefix is /static, it may install
// handlers for /static/first and /static/second on demand.
func NewServer(prefix string, httpServer http.Server) *Server {
    return &Server{prefix: prefix, httpServer: httpServer, installedPaths: make(map[string] string)}
}

func (server *Server) Install(name string, root string) error {
    p := path.Join(server.prefix, name) + "/"
    log.Println("Attempting to install static content server for", root, "at", p)
    if old_root, has := server.installedPaths[p]; has {
        // Path exists. Same?
        if old_root == root {
            log.Println("Server already installed; skipping")
            return nil
        } else {
            err := errors.New("Path collision: name " + name +
                " already installed for " + old_root + "; can't install for " + root)
            log.Println(err)
            return err
        }
    }
    httpRoot := http.Dir(root)
    fileServer := &fileHandler{p, root, httpRoot, http.FileServer(httpRoot)}
	http.Handle(p, http.StripPrefix(p, fileServer))
    server.installedPaths[p] = root
    log.Println("Server installation successful")
    return nil
}

type fileHandler struct {
    PathPrefix string
    OSPath string
    FileRoot http.FileSystem
    fallbackHandler http.Handler
}

func (f *fileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    logRequest(r)
    if r.FormValue("sc_mode") == "raw" {
        f.fallbackHandler.ServeHTTP(w, r)
        return
    }
    if r.FormValue("sc_mode") == "transcode" {
        f.TranscodeAndServe(w, r)
        return
    }
    upath := r.URL.Path
    last := strings.LastIndex(upath, ".")
    if last >= 0 {
        suffix := upath[last+1:]
        switch i := strings.ToLower(suffix); i {
        case "mp4": {
            if *custom_video_player {
                f.ServeVideoPlayer("mp4", false, w, r)
            } else {
                f.fallbackHandler.ServeHTTP(w, r)
            }
            return
        }
        case "mkv","avi","wmv": {
            if *custom_video_player {
                f.ServeVideoPlayer(*transcode_content_type, true, w, r)
            } else {
                f.TranscodeAndServe(w, r)
            }
            return
        }
        }
    }
    
    f.fallbackHandler.ServeHTTP(w, r)
    return
}

type ChannelWriter struct {
    Channel chan []byte
}

func (c *ChannelWriter) Write(b []byte) (n int, e error) {
    defer func() {
        if r := recover(); r != nil {
            n, e = 0, errors.New("Channel closed.")
        }
    }()
    dup := make([]byte, len(b))
    copy(dup, b)
    c.Channel<-append(dup)
    return len(b), nil
}

func (f *fileHandler) TranscodeAndServe(w http.ResponseWriter, r *http.Request) {
    defer func() {
        if r := recover(); r != nil {
            log.Println("Recovered from transcode crash:", r)
        }
    }()
    videoPath := filepath.Clean(filepath.Join(f.OSPath, r.URL.Path))
    rootPath := filepath.Clean(f.OSPath) + string(filepath.Separator)
    if videoPath[:len(rootPath)] != rootPath {
        log.Println("Trying to open path outside filesystem root:", videoPath, "not in", rootPath)
        w.Write([]byte("Error: invalid path."))
        return
    }
    
    transcodeSettings := strings.Split(*transcode_settings, " ")
    args := []string{"-i", videoPath}
    args = append(args, transcodeSettings...)
    cmd := exec.Command(*transcoder, args...)
    platform.Hide(cmd)
    log.Println("Calling", cmd.Path, cmd.Args)

    w.Header().Set("Content-Type", "video/" + *transcode_content_type)
    c := make(chan []byte, 1)
    defer close(c)
    cmd.Stdout = &ChannelWriter{c}
    
    // Redirect all stderr output from the transcoder to the log file.
    if *verbose_transcode_output {
        c2 := make(chan []byte, 1)
        defer close(c2)
        cmd.Stderr = &ChannelWriter{c2}
        go func() {
            for message := range c2 {
                log.Println(string(message))
            }
        }()
    }
    err := cmd.Start()
    done := false
    go func() {
        cmd.Wait()
        done = true
    }()
    if err != nil {
        log.Println("Error:", err)
        w.Write([]byte("Error: " + err.Error()))
        return
    }
    
    for ;; {
        shouldBreak := false
        nothing := true
        select {
        case b, ok := <- c: {
            if ok {
                nothing = false
                if *verbose_transcode_output {
                    log.Println("Read", len(b), "bytes from transcoder")
                }
                n, err := w.Write(b)
                if err != nil {
                    log.Println("Failed to write to output stream. Done!")
                    shouldBreak = true
                } else if n != len(b) {
                    log.Println("Read/Write mismatch; read", len(b), "but wrote", n, ". Aborting stream.")
                    shouldBreak = true
                }
            }
        }
        default: {
        }
        }
        if shouldBreak {
            break
        }
        if (nothing && done) {
            log.Println("File write done; exiting cleanly.")
            break
        } else if (nothing) {
            time.Sleep(10 * time.Millisecond)
        }
    }
    cmd.Process.Kill()
}

var VIDEO_TEMPLATE_FILE = "templates/video.html.template"

type VideoData struct {
    Url string
    DownloadUrl string
    TranscodeUrl string
    Type string
}

func (f *fileHandler) ServeVideoPlayer(t string, transcode bool, w http.ResponseWriter, r *http.Request) {
    template_content, err := ioutil.ReadFile(VIDEO_TEMPLATE_FILE)
    if err != nil {
        f.fallbackHandler.ServeHTTP(w, r)
    }
    
    var videoTemplate = template.New("Video template")
    videoTemplate, err = videoTemplate.Parse(string(template_content))
    if err != nil {
        f.fallbackHandler.ServeHTTP(w, r)
    }
    var videoData *VideoData
    if transcode {
        videoData = &VideoData{Url: "?sc_mode=transcode", DownloadUrl: "?sc_mode=raw",
                               TranscodeUrl: "?sc_mode=transcode", Type: t}
    } else {
        videoData = &VideoData{Url: "?sc_mode=raw", DownloadUrl: "?sc_mode=raw", Type: t}
    }
    videoTemplate.Execute(w, videoData)
}

func (server *Server) Roots() []string {
    roots := make([]string, len(server.installedPaths))
    i := 0
    for k, _ := range server.installedPaths {
        roots[i] = k
        i++
    }
    sort.Strings(roots)
    return roots
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