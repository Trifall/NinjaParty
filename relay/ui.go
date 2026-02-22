package main

import (
	"html/template"
	"net/http"

	"go.uber.org/zap"
)

type UI struct {
	addr        string
	observerURL string
	room        string
	logger      *zap.Logger
}

const uiTemplate = `
<!DOCTYPE html>
<html>
<head>
	<title>Ninjabrain Party Relay</title>
	<style>
		body { font-family: sans-serif; text-align: center; margin-top: 50px; background: #1a1a1a; color: #fff; }
		.box { background: #333; display: inline-block; padding: 20px 40px; border-radius: 8px; }
		a { color: #4CAF50; text-decoration: none; font-weight: bold; }
		a:hover { text-decoration: underline; }
		input { padding: 10px; font-size: 16px; margin: 10px 0; text-align: center; width: 300px; border-radius: 4px; border: 1px solid #555; background: #222; color: #fff; }
	</style>
</head>
<body>
	<div class="box">
		<h2>Ninjabrain Party Relay is running!</h2>
		<p>Your room code is: <strong>{{.RoomCode}}</strong></p>
		<p>Observer Link:</p>
		<input type="text" id="link" value="{{.FullLink}}" readonly onclick="this.select()" />
		<p><a href="{{.FullLink}}" target="_blank">Open Observer locally to test</a></p>
	</div>
</body>
</html>
`

func (u *UI) Start() error {
	tmpl := template.Must(template.New("ui").Parse(uiTemplate))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := struct {
			RoomCode string
			FullLink string
		}{
			RoomCode: u.room,
			FullLink: u.observerURL + u.room,
		}
		if err := tmpl.Execute(w, data); err != nil {
			u.logger.Error("failed to render ui template", zap.Error(err))
		}
	})

	u.logger.Info("starting local UI", zap.String("url", "http://"+u.addr))
	return http.ListenAndServe(u.addr, nil)
}
