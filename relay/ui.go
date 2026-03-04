package main

import (
	_ "embed"
	"html/template"
	"net/http"
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

//go:embed ui.html
var uiTemplateHTML string

type UI struct {
	addr        string
	relay       *Relay
	logger      *zap.Logger
	newRoomRate *RateLimiter
	tmpl        *template.Template
}

type PageData struct {
	RunnerName    string
	RoomCode      string
	FullLink      string
	IsActive      bool
	IsCreator     bool
	StaleRoom     bool
	ShowSetup     bool
	ShowRejoin    bool
	HasRunnerName bool
	SavedRoomCode string
}

type SetupRequest struct {
	RunnerName string `json:"runnerName"`
	Password   string `json:"password,omitempty"`
	RoomCode   string `json:"roomCode,omitempty"`
}

type SetupResponse struct {
	Success  bool   `json:"success"`
	Error    string `json:"error,omitempty"`
	RoomCode string `json:"roomCode,omitempty"`
}

func (u *UI) Start() error {
	u.tmpl = template.Must(template.New("ui").Parse(uiTemplateHTML))
	mux := http.NewServeMux()

	wd, err := os.Getwd()
	if err != nil {
		u.logger.Warn("could not get working directory, static files may not be available", zap.Error(err))
	} else {
		fs := http.FileServer(http.Dir(filepath.Join(wd, "static")))
		mux.Handle("/static/", http.StripPrefix("/static/", fs))
	}

	mux.HandleFunc("/", u.handleIndex)
	mux.HandleFunc("/close", u.handleClose)
	mux.HandleFunc("/force-close", u.handleForceClose)
	mux.HandleFunc("/setup/create", u.handleCreate)
	mux.HandleFunc("/setup/join", u.handleJoin)
	mux.HandleFunc("/setup/rejoin", u.handleRejoin)
	mux.HandleFunc("/status", u.handleStatus)

	u.logger.Info("starting local UI", zap.String("url", "http://"+u.addr))
	return http.ListenAndServe(u.addr, mux)
}
