// +build !windows

package platform

import (
    "errors"
)

// Spawns a window containing a webview pointed at the given URL, and waits for
// it to be closed. If the webview cannot be opened, returns an error
// immediately.
func WebviewWindow(serverURL string) error {
    return errors.New("No webview window available for this platform.")
}