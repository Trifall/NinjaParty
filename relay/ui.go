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
		button { padding: 8px 16px; font-size: 14px; cursor: pointer; border-radius: 4px; border: none; margin-left: 10px; }
		.btn-hide { background: #f44336; color: white; }
		.btn-reveal { background: #4CAF50; color: white; }
		.btn-copy { background: #2196F3; color: white; }
		.dialog-overlay { display: none; position: fixed; top: 0; left: 0; width: 100%; height: 100%; background: rgba(0,0,0,0.7); justify-content: center; align-items: center; }
		.dialog { background: #333; padding: 30px; border-radius: 8px; text-align: center; }
		.dialog-buttons { margin-top: 20px; }
		.dialog-buttons button { margin: 0 10px; }
		.btn-cancel { background: #555; color: white; }
		.btn-confirm { background: #4CAF50; color: white; }
	</style>
</head>
<body>
	<div class="box">
		<h2>Ninjabrain Party Relay is running!</h2>
		<p>Your room code is: <strong id="roomCode">{{.RoomCode}}</strong>
			<button id="toggleBtn" class="btn-hide" onclick="toggleRoomCode()">Hide</button>
		</p>
		<p style="margin-bottom: 5px;">Observer Link:</p>
		<div style="display: flex; justify-content: center; align-items: center; gap: 10px;">
			<input type="text" id="link" value="{{.FullLink}}" readonly onclick="this.select()" />
			<button id="copyBtn" class="btn-copy" onclick="copyLink()">Copy</button>
		</div>
	</div>

	<div id="dialog" class="dialog-overlay">
		<div class="dialog">
			<p>Are you sure you want to reveal the room code?</p>
			<div class="dialog-buttons">
				<button class="btn-cancel" onclick="closeDialog()">Cancel</button>
				<button class="btn-confirm" onclick="confirmReveal()">Reveal</button>
			</div>
		</div>
	</div>

	<script>
		var isHidden = false;
		var originalRoomCode = "{{.RoomCode}}";
		var originalLink = "{{.FullLink}}";

		function toggleRoomCode() {
			if (isHidden) {
				document.getElementById('dialog').style.display = 'flex';
			} else {
				hideRoomCode();
			}
		}

		function confirmReveal() {
			closeDialog();
			showRoomCode();
		}

		function closeDialog() {
			document.getElementById('dialog').style.display = 'none';
		}

		function hideRoomCode() {
			isHidden = true;
			document.getElementById('roomCode').textContent = '******';
			document.getElementById('link').value = '******';
			document.getElementById('testLink').innerHTML = '<span style="color: #888;">(hidden)</span>';
			document.getElementById('toggleBtn').textContent = 'Reveal';
			document.getElementById('toggleBtn').className = 'btn-reveal';
		}

		function showRoomCode() {
			isHidden = false;
			document.getElementById('roomCode').textContent = originalRoomCode;
			document.getElementById('link').value = originalLink;
			document.getElementById('toggleBtn').textContent = 'Hide';
			document.getElementById('toggleBtn').className = 'btn-hide';
		}

		function copyLink() {
			navigator.clipboard.writeText(originalLink).then(function() {
				var btn = document.getElementById('copyBtn');
				btn.textContent = 'Copied!';
				setTimeout(function() { btn.textContent = 'Copy'; }, 2000);
			});
		}
	</script>
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
