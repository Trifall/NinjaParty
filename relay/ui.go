package main

import (
	"html/template"
	"net/http"

	"go.uber.org/zap"
)

type UI struct {
	addr   string
	relay  *Relay
	logger *zap.Logger
}

const uiTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8" />
	<meta name="viewport" content="width=device-width, initial-scale=1.0" />
	<title>Ninjabrain Party Relay</title>
	<style>
		*, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }

		body {
			font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
			background: #0a0a0a;
			color: #fff;
			min-height: 100vh;
			display: flex;
			align-items: center;
			justify-content: center;
			padding: 24px;
		}

		.glow {
			position: fixed;
			top: 50%;
			left: 50%;
			transform: translateX(-50%);
			width: 600px;
			height: 300px;
			background: radial-gradient(ellipse, rgba(16,185,129,0.12) 0%, transparent 70%);
			pointer-events: none;
		}

		.card {
			position: relative;
			background: rgba(255,255,255,0.04);
			border: 1px solid rgba(255,255,255,0.1);
			border-radius: 16px;
			padding: 36px 40px;
			width: 100%;
			max-width: 520px;
			backdrop-filter: blur(20px);
			box-shadow: 0 25px 50px rgba(0,0,0,0.5);
			animation: fadeUp 0.5s ease both;
		}

		@keyframes fadeUp {
			from { opacity: 0; transform: translateY(16px); }
			to   { opacity: 1; transform: translateY(0); }
		}

		.status {
			display: inline-flex;
			align-items: center;
			gap: 7px;
			font-size: 12px;
			font-weight: 500;
			color: #10b981;
			background: rgba(16,185,129,0.1);
			border: 1px solid rgba(16,185,129,0.2);
			border-radius: 999px;
			padding: 4px 12px;
			margin-bottom: 20px;
			letter-spacing: 0.04em;
			text-transform: uppercase;
		}

		.status-dot {
			width: 6px;
			height: 6px;
			background: #10b981;
			border-radius: 50%;
			animation: pulse 2s ease infinite;
		}

		@keyframes pulse {
			0%, 100% { opacity: 1; }
			50% { opacity: 0.3; }
		}

		h1 {
			font-size: 22px;
			font-weight: 700;
			color: #fff;
			margin-bottom: 4px;
			letter-spacing: -0.02em;
		}

		.subtitle {
			font-size: 13px;
			color: rgba(255,255,255,0.75);
			margin-bottom: 28px;
		}

		.divider {
			border: none;
			border-top: 1px solid rgba(255,255,255,0.07);
			margin-bottom: 24px;
		}

		.field-label {
			font-size: 16px;
			font-weight: 600;
			letter-spacing: 0.06em;
			text-transform: uppercase;
			color: rgba(255,255,255,0.75);
			margin-bottom: 8px;
		}

		.room-row {
			display: flex;
			align-items: center;
			justify-content: space-between;
			background: rgba(255,255,255,0.04);
			border: 1px solid rgba(255,255,255,0.08);
			border-radius: 10px;
			padding: 12px 16px;
			margin-bottom: 16px;
		}

		#roomCode {
			font-size: 22px;
			font-weight: 800;
			letter-spacing: 0.12em;
			color: #fff;
			font-variant-numeric: tabular-nums;
		}

		.link-row {
			display: flex;
			gap: 8px;
			margin-bottom: 0;
		}

		input[type="text"] {
			flex: 1;
			background: rgba(255,255,255,0.04);
			border: 1px solid rgba(255,255,255,0.08);
			border-radius: 10px;
			padding: 11px 14px;
			font-size: 13px;
			color: rgba(255,255,255,0.8);
			outline: none;
			cursor: pointer;
			transition: border-color 0.2s;
			min-width: 0;
		}

		input[type="text"]:focus {
			border-color: rgba(16,185,129,0.4);
		}

		button {
			display: inline-flex;
			align-items: center;
			gap: 6px;
			padding: 10px 16px;
			font-size: 13px;
			font-weight: 500;
			border: none;
			border-radius: 10px;
			cursor: pointer;
			transition: opacity 0.15s, transform 0.1s;
			white-space: nowrap;
		}

		button:active { transform: scale(0.97); }

		.btn-toggle {
			background: rgba(255,255,255,0.07);
			color: rgba(255,255,255,0.7);
			border: 1px solid rgba(255,255,255,0.1);
		}

		.btn-toggle:hover { background: rgba(255,255,255,0.12); }

		.btn-copy {
			background: #10b981;
			color: #fff;
		}

		.btn-copy:hover { opacity: 0.88; }

		/* Dialog */
		.dialog-overlay {
			display: none;
			position: fixed;
			inset: 0;
			background: rgba(0,0,0,0.7);
			backdrop-filter: blur(4px);
			justify-content: center;
			align-items: center;
			z-index: 100;
		}

		.dialog {
			background: #1a1a1a;
			border: 1px solid rgba(255,255,255,0.1);
			border-radius: 14px;
			padding: 28px 32px;
			text-align: center;
			max-width: 340px;
			width: 100%;
			box-shadow: 0 20px 40px rgba(0,0,0,0.6);
			animation: fadeUp 0.2s ease both;
		}

		.dialog p {
			font-size: 15px;
			color: rgba(255,255,255,0.8);
			margin-bottom: 20px;
			line-height: 1.5;
		}

		.dialog-buttons {
			display: flex;
			gap: 10px;
			justify-content: center;
		}

		.btn-cancel {
			background: rgba(255,255,255,0.07);
			color: rgba(255,255,255,0.6);
			border: 1px solid rgba(255,255,255,0.1);
			padding: 9px 20px;
		}

		.btn-confirm {
			background: #10b981;
			color: #fff;
			padding: 9px 20px;
		}

		.btn-close {
			background: #dc2626;
			color: #fff;
			margin-top: 28px;
			width: 100%;
			padding: 12px;
			font-size: 14px;
			font-weight: 600;
			justify-content: center;
		}

		.btn-close:hover { background: #ef4444; }

		.instructions-container {
			margin-top: 32px;
			display: flex;
			flex-direction: column;
			gap: 10px;
		}

		.instruction {
			font-size: 16px;
			color: rgba(255,255,255,0.8);
			line-height: 1.5;
			display: flex;
			gap: 10px;
			text-align: left;
			letter-spacing: -0.01em;
		}

		.instruction::before {
			content: "•";
			color: #10b981;
			font-weight: 800;
		}
	</style>
</head>
<body>
	<div class="glow"></div>

	<div class="card">
		<div class="status">
			<span class="status-dot"></span>
			Relay Active
		</div>

		<h1>Ninjabrain Party</h1>
		<p class="subtitle">Share the link below with your observer to get started.</p>

		<hr class="divider" />

		<p class="field-label">Room Code</p>
		<div class="room-row">
			<span id="roomCode">{{.RoomCode}}</span>
			<button id="toggleBtn" class="btn-toggle" onclick="toggleRoomCode()">Hide</button>
		</div>

		<p class="field-label">Observer Link</p>
		<div class="link-row">
			<input type="text" id="link" value="{{.FullLink}}" readonly onclick="this.select()" />
			<button id="copyBtn" class="btn-copy" onclick="copyLink()">Copy</button>
		</div>

		<div class="instructions-container">
			<p class="instruction">Once you give out the code / link, you can close this window.</p>
			<p class="instruction">You can open this page again manually or by clicking the tray icon.</p>
			<p class="instruction">If you quit via the tray icon, everything will be closed for you!</p>
		</div>
		<button class="btn-close" onclick="showCloseDialog()">CLOSE RELAY</button>
	</div>

	<div id="dialog" class="dialog-overlay">
		<div class="dialog">
			<p>Reveal the room code and observer link?</p>
			<div class="dialog-buttons">
				<button class="btn-cancel" onclick="closeDialog()">Cancel</button>
				<button class="btn-confirm" onclick="confirmReveal()">Reveal</button>
			</div>
		</div>
	</div>

	<div id="closeDialog" class="dialog-overlay">
		<div class="dialog">
			<p>Are you sure you want to close the relay? This will disconnect all observers.</p>
			<div class="dialog-buttons">
				<button class="btn-cancel" onclick="closeCloseDialog()">Cancel</button>
				<button class="btn-confirm" style="background: #dc2626;" onclick="confirmClose()">Close Relay</button>
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
			document.getElementById('roomCode').textContent = '••••••';
			document.getElementById('link').value = '••••••';
			document.getElementById('toggleBtn').textContent = 'Reveal';
		}

		function showRoomCode() {
			isHidden = false;
			document.getElementById('roomCode').textContent = originalRoomCode;
			document.getElementById('link').value = originalLink;
			document.getElementById('toggleBtn').textContent = 'Hide';
		}

		function copyLink() {
			navigator.clipboard.writeText(originalLink).then(function() {
				var btn = document.getElementById('copyBtn');
				btn.textContent = 'Copied!';
				setTimeout(function() { btn.textContent = 'Copy'; }, 2000);
			});
		}

		function showCloseDialog() {
			document.getElementById('closeDialog').style.display = 'flex';
		}

		function closeCloseDialog() {
			document.getElementById('closeDialog').style.display = 'none';
		}

		function confirmClose() {
			closeCloseDialog();
			document.querySelector('.card').style.display = 'none';
			document.querySelector('.glow').style.display = 'none';
			document.body.innerHTML = 
				'<div style="display:flex;flex-direction:column;align-items:center;justify-content:center;height:100vh;color:rgba(255,255,255,0.9);font-size:20px;gap:20px;">' +
					'<div>Relay closed.</div>' +
					'<button id="newRoomBtn" style="background:#10b981;color:#fff;border:none;border-radius:10px;padding:12px 24px;font-size:14px;font-weight:500;cursor:pointer;" onclick="startNewRoom()">Start New Room</button>' +
				'</div>';
			fetch('/close', { method: 'POST' });
		}

		function startNewRoom() {
			window.location.href = '/new-room';
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
			RoomCode: u.relay.RoomCode(),
			FullLink: u.relay.FullLink(),
		}
		if err := tmpl.Execute(w, data); err != nil {
			u.logger.Error("failed to render ui template", zap.Error(err))
		}
	})

	http.HandleFunc("/close", func(w http.ResponseWriter, r *http.Request) {
		u.logger.Info("close requested via UI")
		u.relay.Shutdown()
	})

	http.HandleFunc("/new-room", func(w http.ResponseWriter, r *http.Request) {
		u.logger.Info("new room requested via UI")
		_, err := u.relay.CreateNewRoom()
		if err != nil {
			u.logger.Error("failed to create new room", zap.Error(err))
			http.Error(w, "Failed to create new room", http.StatusInternalServerError)
			return
		}
		// Redirect to home page to show the new room
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	u.logger.Info("starting local UI", zap.String("url", "http://"+u.addr))
	return http.ListenAndServe(u.addr, nil)
}
