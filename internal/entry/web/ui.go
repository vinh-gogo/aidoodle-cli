package web

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed ui/*
var embeddedUI embed.FS

// FileSystem trả về http.FileSystem phục vụ các tệp giao diện web tĩnh.
func FileSystem() http.FileSystem {
	sub, err := fs.Sub(embeddedUI, "ui")
	if err != nil {
		panic(err)
	}
	return http.FS(sub)
}

// IndexHTML đọc trực tiếp tệp index.html đã nhúng.
func IndexHTML() ([]byte, error) {
	return embeddedUI.ReadFile("ui/index.html")
}
