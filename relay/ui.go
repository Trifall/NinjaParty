package main

import (
	"encoding/json"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
)

type RateLimiter struct {
	mu          sync.Mutex
	lastRequest time.Time
	minInterval time.Duration
}

func NewRateLimiter(interval time.Duration) *RateLimiter {
	return &RateLimiter{
		minInterval: interval,
	}
}

func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	if now.Sub(rl.lastRequest) < rl.minInterval {
		return false
	}
	rl.lastRequest = now
	return true
}

type UI struct {
	addr        string
	relay       *Relay
	logger      *zap.Logger
	newRoomRate *RateLimiter
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

const uiTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8" />
	<meta name="viewport" content="width=device-width, initial-scale=1.0" />
	<title>NinjaParty Relay</title>
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

		.status-disconnected {
			color: #f59e0b;
			background: rgba(245,158,11,0.1);
			border-color: rgba(245,158,11,0.2);
		}

		.status-disconnected .status-dot {
			background: #f59e0b;
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

		input[type="text"], input[type="password"] {
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

		input[type="text"]:focus, input[type="password"]:focus {
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

		.btn-primary {
			background: #10b981;
			color: #fff;
		}

		.btn-primary:hover { opacity: 0.88; }

		.btn-secondary {
			background: rgba(255,255,255,0.07);
			color: rgba(255,255,255,0.7);
			border: 1px solid rgba(255,255,255,0.1);
		}

		.btn-secondary:hover { background: rgba(255,255,255,0.12); }

		.btn-danger {
			background: #dc2626;
			color: #fff;
		}

		.btn-danger:hover { background: #ef4444; }

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

		.error {
			color: #ef4444;
			font-size: 13px;
			margin-top: 12px;
			text-align: center;
		}

		.success {
			color: #10b981;
			font-size: 13px;
			margin-top: 12px;
			text-align: center;
		}

		.tabs {
			display: flex;
			gap: 8px;
			margin-bottom: 24px;
		}

		.tab {
			flex: 1;
			padding: 12px;
			text-align: center;
			background: rgba(255,255,255,0.04);
			border: 1px solid rgba(255,255,255,0.08);
			border-radius: 10px;
			color: rgba(255,255,255,0.6);
			cursor: pointer;
			transition: all 0.2s;
		}

		.tab.active {
			background: rgba(16,185,129,0.15);
			border-color: rgba(16,185,129,0.4);
			color: #10b981;
		}

		.tab:hover:not(.active) {
			background: rgba(255,255,255,0.08);
		}

		.panel {
			display: none;
		}

		.panel.active {
			display: block;
		}

		.form-group {
			margin-bottom: 16px;
		}

		.runner-badge {
			display: inline-flex;
			align-items: center;
			gap: 6px;
			background: rgba(16,185,129,0.15);
			border: 1px solid rgba(16,185,129,0.3);
			border-radius: 8px;
			padding: 6px 12px;
			font-size: 14px;
			color: #10b981;
			margin-bottom: 16px;
		}

		.instructions-container {
			margin-top: 32px;
			display: flex;
			flex-direction: column;
			gap: 10px;
		}

		.instruction {
			font-size: 14px;
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

		.btn-help {
			margin-top: 16px;
			width: 100%;
			background: rgba(16,185,129,0.15);
			color: #10b981;
			border: 1px solid rgba(16,185,129,0.3);
			padding: 10px;
			font-size: 12px;
			font-weight: 500;
		}

		.btn-help:hover {
			background: rgba(16,185,129,0.25);
			color: #34d399;
		}

		.btn-close {
			margin-top: 28px;
			width: 100%;
			padding: 12px;
			font-size: 14px;
			font-weight: 600;
			justify-content: center;
		}

		.stale-room-banner {
			background: rgba(245,158,11,0.15);
			border: 1px solid rgba(245,158,11,0.3);
			border-radius: 10px;
			padding: 16px;
			margin-bottom: 20px;
			text-align: center;
			color: #f59e0b;
		}

		.help-content {
			text-align: left;
		}

		.help-content h2 {
			font-size: 18px;
			font-weight: 600;
			color: #fff;
			margin-bottom: 16px;
			text-align: center;
		}

		.help-content ol {
			margin: 0;
			padding-left: 20px;
			color: rgba(255,255,255,0.8);
			font-size: 14px;
			line-height: 1.8;
		}

		.help-content li {
			margin-bottom: 8px;
		}

		.help-content img {
			display: block;
			max-width: 100%;
			border-radius: 8px;
			margin: 16px auto;
			border: 1px solid rgba(255,255,255,0.1);
		}

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
	</style>
</head>
<body>
	<div class="glow"></div>

	{{if .ShowSetup}}
	<div class="card">
		<h1>NinjaParty Setup</h1>
		<p class="subtitle">Enter your runner name to get started</p>

		<div class="form-group">
			<p class="field-label">Runner Name</p>
			<input type="text" id="runnerName" placeholder="Your name (max 24 chars)" maxlength="24" value="{{.RunnerName}}" />
		</div>

		<div class="tabs">
			<div class="tab active" onclick="switchTab('create')" id="tabCreate">Create Room</div>
			<div class="tab" onclick="switchTab('join')" id="tabJoin">Join Room</div>
		</div>

		<div class="panel active" id="panelCreate">
			<div class="form-group">
				<input type="password" id="createPassword" placeholder="Room password (optional)" />
			</div>
			<button class="btn-primary" style="width:100%" onclick="createRoom()">Create Room</button>
		</div>

		<div class="panel" id="panelJoin">
			<div class="form-group">
				<input type="text" id="joinRoomCode" placeholder="Room code" maxlength="6" />
			</div>
			<div class="form-group">
				<input type="password" id="joinPassword" placeholder="Password (if required)" />
			</div>
			<button class="btn-primary" style="width:100%" onclick="joinRoom()">Join Room</button>
		</div>

		<div id="setupError" class="error"></div>
	</div>
	{{else if .ShowRejoin}}
	<div class="card">
		<h1>NinjaParty</h1>
		<p class="subtitle">Your previous room is still active</p>

		<div class="runner-badge">
			<span>Runner: {{.RunnerName}}</span>
		</div>

		<div style="background: rgba(16,185,129,0.15); border: 1px solid rgba(16,185,129,0.3); border-radius: 10px; padding: 16px; margin-bottom: 20px; text-align: center; color: #10b981;">
			Room <strong>{{.SavedRoomCode}}</strong> is still active. You can rejoin or start fresh.
		</div>

		<button class="btn-primary" style="width:100%; margin-bottom: 12px;" onclick="rejoinRoom()">Rejoin Room {{.SavedRoomCode}}</button>

		<div class="tabs" style="margin-top: 24px;">
			<div class="tab active" onclick="switchTab('create')" id="tabCreate">Create New Room</div>
			<div class="tab" onclick="switchTab('join')" id="tabJoin">Join Different Room</div>
		</div>

		<div class="panel active" id="panelCreate">
			<div class="form-group">
				<input type="password" id="createPassword" placeholder="Room password (optional)" />
			</div>
			<button class="btn-primary" style="width:100%" onclick="createRoom()">Create Room</button>
		</div>

		<div class="panel" id="panelJoin">
			<div class="form-group">
				<input type="text" id="joinRoomCode" placeholder="Room code" maxlength="6" />
			</div>
			<div class="form-group">
				<input type="password" id="joinPassword" placeholder="Password (if required)" />
			</div>
			<button class="btn-primary" style="width:100%" onclick="joinRoom()">Join Room</button>
		</div>

		<div id="setupError" class="error"></div>
	</div>
	{{else if .StaleRoom}}
	<div class="card">
		<h1>NinjaParty</h1>
		<p class="subtitle">Your previous room no longer exists</p>

		<div class="stale-room-banner">
			The room you were in has expired or been closed.
		</div>

		<div class="tabs">
			<div class="tab active" onclick="switchTab('create')" id="tabCreate">Create Room</div>
			<div class="tab" onclick="switchTab('join')" id="tabJoin">Join Room</div>
		</div>

		<div class="panel active" id="panelCreate">
			<div class="form-group">
				<input type="password" id="createPassword" placeholder="Room password (optional)" />
			</div>
			<button class="btn-primary" style="width:100%" onclick="createRoom()">Create Room</button>
		</div>

		<div class="panel" id="panelJoin">
			<div class="form-group">
				<input type="text" id="joinRoomCode" placeholder="Room code" maxlength="6" />
			</div>
			<div class="form-group">
				<input type="password" id="joinPassword" placeholder="Password (if required)" />
			</div>
			<button class="btn-primary" style="width:100%" onclick="joinRoom()">Join Room</button>
		</div>

		<div id="setupError" class="error"></div>
	</div>
	{{else}}
	<div class="card">
		<div class="status">
			<span class="status-dot"></span>
			Relay Active
		</div>
		<div id="ninjabrainStatus" class="status status-disconnected">
			<span class="status-dot"></span>
			<span id="ninjabrainStatusText">Waiting for Ninjabrain...</span>
		</div>

		<div class="runner-badge">
			<span>Runner: {{.RunnerName}}</span>
		</div>

		<h1>NinjaParty</h1>
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
		<button class="btn-help" onclick="showApiHelp()">How do I enable the Ninjabrain API?</button>
		{{if .IsCreator}}
		<button class="btn-danger btn-close" onclick="showForceCloseDialog()">FORCE CLOSE ROOM</button>
		{{else}}
		<button class="btn-close" onclick="showCloseDialog()">LEAVE ROOM</button>
		{{end}}
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
			<p>Are you sure you want to leave the room? Observers will still be able to see the last known data.</p>
			<div class="dialog-buttons">
				<button class="btn-cancel" onclick="closeCloseDialog()">Cancel</button>
				<button class="btn-confirm" onclick="confirmClose()">Leave Room</button>
			</div>
		</div>
	</div>

	<div id="forceCloseDialog" class="dialog-overlay">
		<div class="dialog">
			<p>Are you sure you want to force close the room? This will disconnect all observers and all runners.</p>
			<div class="dialog-buttons">
				<button class="btn-cancel" onclick="closeForceCloseDialog()">Cancel</button>
				<button class="btn-confirm" style="background:#dc2626" onclick="confirmForceClose()">Force Close</button>
			</div>
		</div>
	</div>

	<div id="apiHelpDialog" class="dialog-overlay">
		<div class="dialog" style="max-width: 420px;">
			<div class="help-content">
				<h2>Enable Ninjabrain API</h2>
				<ol>
					<li>Open Ninjabrain Bot</li>
					<li>Go to <strong>Settings</strong> → <strong>Advanced</strong> tab</li>
					<li>Check <strong>"Enable API (starts HTTP server)"</strong></li>
				</ol>
				<img src="/static/ninjabrain_enable_api_setting.png" alt="Enable API setting in Ninjabrain Bot" />
			</div>
			<button class="btn-close" onclick="closeApiHelp()">Got it</button>
		</div>
	</div>
	{{end}}

	<script>
		var isHidden = false;
		var originalRoomCode = "{{.RoomCode}}";
		var originalLink = "{{.FullLink}}";
		var savedRoomCode = "{{.SavedRoomCode}}";

		{{if or .ShowSetup .StaleRoom .ShowRejoin}}
		function switchTab(tab) {
			document.getElementById('tabCreate').classList.toggle('active', tab === 'create');
			document.getElementById('tabJoin').classList.toggle('active', tab === 'join');
			document.getElementById('panelCreate').classList.toggle('active', tab === 'create');
			document.getElementById('panelJoin').classList.toggle('active', tab === 'join');
			document.getElementById('setupError').textContent = '';
		}

		async function rejoinRoom() {
			var errorEl = document.getElementById('setupError');
			errorEl.textContent = 'Rejoining room...';

			var res = await fetch('/setup/rejoin', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ roomCode: savedRoomCode })
			});

			var data = await res.json();
			if (!res.ok) {
				errorEl.textContent = data.error || 'Failed to rejoin room';
				return;
			}

			location.reload();
		}

		async function createRoom() {
			var name = document.getElementById('runnerName').value.trim();
			var password = document.getElementById('createPassword').value;
			var errorEl = document.getElementById('setupError');

			if (!name) {
				errorEl.textContent = 'Please enter a runner name';
				return;
			}

			errorEl.textContent = 'Creating room...';

			var res = await fetch('/setup/create', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ runnerName: name, password: password })
			});

			var data = await res.json();
			if (!res.ok) {
				errorEl.textContent = data.error || 'Failed to create room';
				return;
			}

			location.reload();
		}

		async function joinRoom() {
			var name = document.getElementById('runnerName').value.trim();
			var code = document.getElementById('joinRoomCode').value.trim();
			var password = document.getElementById('joinPassword').value;
			var errorEl = document.getElementById('setupError');

			if (!name) {
				errorEl.textContent = 'Please enter a runner name';
				return;
			}
			if (!code) {
				errorEl.textContent = 'Please enter a room code';
				return;
			}

			errorEl.textContent = 'Joining room...';

			var res = await fetch('/setup/join', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ runnerName: name, roomCode: code, password: password })
			});

			var data = await res.json();
			if (!res.ok) {
				errorEl.textContent = data.error || 'Failed to join room';
				return;
			}

			location.reload();
		}
		{{else}}
		function updateNinjabrainStatus() {
			fetch('/status')
				.then(function(response) { return response.json(); })
				.then(function(data) {
					var statusEl = document.getElementById('ninjabrainStatus');
					var statusText = document.getElementById('ninjabrainStatusText');
					if (data.ninjabrainConnected) {
						statusEl.classList.remove('status-disconnected');
						statusText.textContent = 'Ninjabrain API Connected';
					} else {
						statusEl.classList.add('status-disconnected');
						statusText.textContent = 'Waiting for Ninjabrain...';
					}
				})
				.catch(function() {
				});
		}

		setInterval(updateNinjabrainStatus, 2000);
		updateNinjabrainStatus();

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
			fetch('/close', { method: 'POST' });
			location.href = '/';
		}

		function showForceCloseDialog() {
			document.getElementById('forceCloseDialog').style.display = 'flex';
		}

		function closeForceCloseDialog() {
			document.getElementById('forceCloseDialog').style.display = 'none';
		}

		function confirmForceClose() {
			closeForceCloseDialog();
			fetch('/force-close', { method: 'POST' });
			location.href = '/';
		}

		function showApiHelp() {
			document.getElementById('apiHelpDialog').style.display = 'flex';
		}

		function closeApiHelp() {
			document.getElementById('apiHelpDialog').style.display = 'none';
		}
		{{end}}
	</script>
</body>
</html>
`

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
	tmpl := template.Must(template.New("ui").Parse(uiTemplate))

	wd, err := os.Getwd()
	if err != nil {
		u.logger.Warn("could not get working directory, static files may not be available", zap.Error(err))
	} else {
		fs := http.FileServer(http.Dir(filepath.Join(wd, "static")))
		http.Handle("/static/", http.StripPrefix("/static/", fs))
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
		w.Header().Set("Pragma", "no-cache")

		state, err := loadRelayState()
		if err != nil {
			state = &RelayState{}
		}

		isActive := u.relay.RoomCode() != ""
		roomCode := u.relay.RoomCode()
		fullLink := u.relay.FullLink()

		showSetup := (state.RunnerName == "" || state.RoomCode == "") && !isActive
		staleRoom := false
		showRejoin := false

		if state.RoomCode != "" && !isActive && state.RunnerName != "" {
			exists := roomExists(u.relay.ServerAPIURL(), state.RoomCode)
			if exists {
				showRejoin = true
			} else {
				staleRoom = true
			}
		}

		data := PageData{
			RunnerName:    state.RunnerName,
			RoomCode:      roomCode,
			FullLink:      fullLink,
			IsActive:      isActive,
			IsCreator:     u.relay.IsCreator(),
			StaleRoom:     staleRoom,
			ShowSetup:     showSetup,
			ShowRejoin:    showRejoin,
			HasRunnerName: state.RunnerName != "",
			SavedRoomCode: state.RoomCode,
		}

		if err := tmpl.Execute(w, data); err != nil {
			u.logger.Error("failed to render ui template", zap.Error(err))
		}
	})

	http.HandleFunc("/close", func(w http.ResponseWriter, r *http.Request) {
		u.logger.Info("close requested via UI")
		u.relay.Shutdown()
		state, _ := loadRelayState()
		state.RoomCode = ""
		state.CreatorToken = ""
		saveRelayState(state)
	})

	http.HandleFunc("/force-close", func(w http.ResponseWriter, r *http.Request) {
		u.logger.Info("force close requested via UI")
		u.relay.Shutdown()
		clearRelayState()
	})

	http.HandleFunc("/setup/create", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if !u.newRoomRate.Allow() {
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(SetupResponse{
				Success: false,
				Error:   "rate limit exceeded, please wait before creating another room",
			})
			return
		}

		var req SetupRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(SetupResponse{
				Success: false,
				Error:   "invalid request",
			})
			return
		}

		if req.RunnerName == "" {
			json.NewEncoder(w).Encode(SetupResponse{
				Success: false,
				Error:   "runner name is required",
			})
			return
		}

		u.relay.SetRunnerName(req.RunnerName)

		roomRes, err := u.relay.CreateNewRoom(req.Password)
		if err != nil {
			u.logger.Error("failed to create room", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(SetupResponse{
				Success: false,
				Error:   "Failed to create room: " + err.Error(),
			})
			return
		}

		state := &RelayState{
			RunnerName:   req.RunnerName,
			RoomCode:     roomRes.RoomCode,
			CreatorToken: roomRes.CreatorToken,
		}
		saveRelayState(state)

		json.NewEncoder(w).Encode(SetupResponse{
			Success:  true,
			RoomCode: roomRes.RoomCode,
		})
	})

	http.HandleFunc("/setup/join", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var req SetupRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(SetupResponse{
				Success: false,
				Error:   "invalid request",
			})
			return
		}

		if req.RunnerName == "" {
			json.NewEncoder(w).Encode(SetupResponse{
				Success: false,
				Error:   "runner name is required",
			})
			return
		}

		if req.RoomCode == "" {
			json.NewEncoder(w).Encode(SetupResponse{
				Success: false,
				Error:   "room code is required",
			})
			return
		}

		u.relay.SetRunnerName(req.RunnerName)

		roomRes, err := u.relay.JoinRoom(req.RoomCode, req.Password)
		if err != nil {
			u.logger.Error("failed to join room", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(SetupResponse{
				Success: false,
				Error:   err.Error(),
			})
			return
		}

		state := &RelayState{
			RunnerName: req.RunnerName,
			RoomCode:   roomRes.RoomCode,
		}
		saveRelayState(state)

		json.NewEncoder(w).Encode(SetupResponse{
			Success:  true,
			RoomCode: roomRes.RoomCode,
		})
	})

	http.HandleFunc("/setup/rejoin", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var req SetupRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(SetupResponse{
				Success: false,
				Error:   "invalid request",
			})
			return
		}

		if req.RoomCode == "" {
			json.NewEncoder(w).Encode(SetupResponse{
				Success: false,
				Error:   "room code is required",
			})
			return
		}

		state, err := loadRelayState()
		if err != nil || state.RunnerName == "" {
			json.NewEncoder(w).Encode(SetupResponse{
				Success: false,
				Error:   "no saved runner name found",
			})
			return
		}

		// Try to rejoin with saved credentials
		var roomRes *RoomResponse
		if state.CreatorToken != "" {
			// Try as creator first
			roomRes, err = u.relay.JoinRoom(req.RoomCode, "")
		} else {
			// Try as regular member
			roomRes, err = u.relay.JoinRoom(req.RoomCode, "")
		}

		if err != nil {
			u.logger.Error("failed to rejoin room", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(SetupResponse{
				Success: false,
				Error:   err.Error(),
			})
			return
		}

		// Update state with new tokens
		state.RoomCode = roomRes.RoomCode
		if roomRes.CreatorToken != "" {
			state.CreatorToken = roomRes.CreatorToken
		}
		saveRelayState(state)

		json.NewEncoder(w).Encode(SetupResponse{
			Success:  true,
			RoomCode: roomRes.RoomCode,
		})
	})

	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{
			"ninjabrainConnected": u.relay.NinjabrainConnected(),
		})
	})

	u.logger.Info("starting local UI", zap.String("url", "http://"+u.addr))
	return http.ListenAndServe(u.addr, nil)
}
