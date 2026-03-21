package ui

import (
	_ "embed"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"

	"github.com/NinjaParty/relay/core"
)

//go:embed ui.html
var uiTemplateHTML string

type UI struct {
	addr        string
	relay       *core.Relay
	logger      *zap.Logger
	newRoomRate *core.RateLimiter
	tmpl        *template.Template
	version     string
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
	Version       string
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

func New(addr string, relay *core.Relay, logger *zap.Logger, newRoomRateInterval time.Duration, version string) *UI {
	return &UI{
		addr:        addr,
		relay:       relay,
		newRoomRate: core.NewRateLimiter(newRoomRateInterval),
		logger:      logger,
		version:     version,
	}
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
