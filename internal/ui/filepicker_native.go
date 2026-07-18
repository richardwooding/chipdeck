//go:build !js

package ui

// Native builds rely on drag-and-drop; no file dialog.
func canPickFiles() bool                     { return false }
func openFilePicker()                        {}
func takePickedFile() ([]byte, string, bool) { return nil, "", false }
