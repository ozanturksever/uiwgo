//go:build js && wasm

package main

import (
	"fmt"
	"syscall/js"

	"github.com/ozanturksever/logutil"
	comps "github.com/ozanturksever/uiwgo/comps"
	dom "github.com/ozanturksever/uiwgo/dom"
	reactivity "github.com/ozanturksever/uiwgo/reactivity"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type FileInfo struct {
	Name string
	Size int64
	Type string
}

func FileUploadExample() Node {
	uploadedFiles := reactivity.CreateSignal([]FileInfo{})
	dragOver := reactivity.CreateSignal(false)

	return Div(
		Style("max-width: 800px; margin: 0 auto; padding: 20px; font-family: Arial, sans-serif;"),
		H1(Text("File Upload Example")),
		P(Text("This example demonstrates file upload using OnFileSelectInline and OnFileDropInline handlers.")),

		// Drop zone
		Div(
			Style("border: 3px dashed #ccc; padding: 40px; text-align: center; margin: 20px 0; border-radius: 8px; cursor: pointer; transition: all 0.3s ease;"),
			Class("drop-zone"),
			Div(
				H3(Text("📁 Drop Files Here")),
				P(Text("or click to select files")),
				P(
					Style("color: #666; font-size: 14px;"),
					Text("Supports multiple files"),
				),
			),
			dom.OnFileDropInline(func(el dom.Element, files []js.Value) {
				dragOver.Set(false)
				processFiles(files, uploadedFiles)
			}),
			dom.OnClickInline(func(el dom.Element) {
				// Create and trigger hidden file input
				input := dom.Document.CreateElement("input")
				input.SetAttribute("type", "file")
				input.SetAttribute("multiple", "true")
				input.SetAttribute("accept", "*/*")
				
				// Add change handler
				input.AddEventListener("change", false, func(event dom.Event) {
					files := event.Target().Underlying().Get("files")
					if !files.IsUndefined() && files.Length() > 0 {
						fileArray := make([]js.Value, files.Length())
						for i := 0; i < files.Length(); i++ {
							fileArray[i] = files.Index(i)
						}
						processFiles(fileArray, uploadedFiles)
					}
				})
				
				input.Underlying().Call("click")
			}),
			// Drag over handlers for visual feedback
			dom.OnDragOverInline(func(el dom.Element, dataTransfer js.Value) {
				dragOver.Set(true)
			}),
		),

		// Traditional file input (alternative method)
		Div(
			Style("margin: 20px 0;"),
			H3(Text("Alternative: File Input")),
			Input(
				Type("file"),
				Multiple(),
				Accept("*/*"),
				Style("padding: 10px; border: 1px solid #ccc; border-radius: 4px;"),
				dom.OnFileSelectInline(func(el dom.Element, files []js.Value) {
					processFiles(files, uploadedFiles)
				}),
			),
		),

		// File list
		Div(
			Style("margin-top: 30px;"),
			H3(Text("Uploaded Files")),
			comps.BindHTML(func() Node {
				files := uploadedFiles.Get()
				if len(files) == 0 {
					return P(
						Style("color: #666; font-style: italic;"),
						Text("No files uploaded yet"),
					)
				}

				items := make([]Node, len(files))
				for i, file := range files {
					items[i] = Div(
						Style("border: 1px solid #ddd; padding: 15px; margin: 10px 0; border-radius: 4px; background: #f9f9f9;"),
						Div(
							Style("display: flex; justify-content: space-between; align-items: center;"),
							Div(
								H4(
									Style("margin: 0 0 5px 0; color: #333;"),
									Text(file.Name),
								),
								P(
									Style("margin: 0; color: #666; font-size: 14px;"),
									Text(fmt.Sprintf("Size: %s | Type: %s", formatFileSize(file.Size), file.Type)),
								),
							),
							Button(
								Style("background: #dc3545; color: white; border: none; padding: 5px 10px; border-radius: 3px; cursor: pointer;"),
								Text("Remove"),
								dom.OnClickInline(func(el dom.Element) {
									removeFile(i, uploadedFiles)
								}),
							),
						),
					)
				}

				return Div(items...)
			}),
		),

		// Clear all button
		comps.BindHTML(func() Node {
			if len(uploadedFiles.Get()) == 0 {
				return Text("")
			}
			return Button(
				Style("background: #6c757d; color: white; border: none; padding: 10px 20px; border-radius: 4px; cursor: pointer; margin-top: 20px;"),
				Text("Clear All Files"),
				dom.OnClickInline(func(el dom.Element) {
					uploadedFiles.Set([]FileInfo{})
				}),
			)
		}),
	)
}

func processFiles(files []js.Value, uploadedFiles reactivity.Signal[[]FileInfo]) {
	currentFiles := uploadedFiles.Get()
	newFiles := make([]FileInfo, 0, len(files))

	for _, file := range files {
		fileInfo := FileInfo{
			Name: file.Get("name").String(),
			Size: int64(file.Get("size").Int()),
			Type: file.Get("type").String(),
		}
		newFiles = append(newFiles, fileInfo)
		logutil.Logf("Processing file: %s (%s)", fileInfo.Name, formatFileSize(fileInfo.Size))
	}

	// Append new files to existing ones
	allFiles := append(currentFiles, newFiles...)
	uploadedFiles.Set(allFiles)
}

func removeFile(index int, uploadedFiles reactivity.Signal[[]FileInfo]) {
	currentFiles := uploadedFiles.Get()
	if index >= 0 && index < len(currentFiles) {
		newFiles := make([]FileInfo, 0, len(currentFiles)-1)
		newFiles = append(newFiles, currentFiles[:index]...)
		newFiles = append(newFiles, currentFiles[index+1:]...)
		uploadedFiles.Set(newFiles)
	}
}

func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func main() {
	comps.Mount("app", func() comps.Node {
		return FileUploadExample()
	})
}