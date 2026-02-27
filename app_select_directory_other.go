//go:build !linux

package main

import wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"

const selectDirectoryDialogTitle = "Select Git Repository"

// SelectDirectory opens a native OS folder picker and returns the selected path.
// Returns an empty string if the user cancels or an error occurs.
func (a *App) SelectDirectory() string {
	path, err := wailsRuntime.OpenDirectoryDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: selectDirectoryDialogTitle,
	})
	if err != nil {
		return ""
	}
	return path
}
