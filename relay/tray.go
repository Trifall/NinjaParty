package main

import (
	_ "embed"
	"fmt"
	"os"
	"time"

	"github.com/atotto/clipboard"
	"github.com/getlantern/systray"
	"github.com/pkg/browser"
	"go.uber.org/zap"

	"github.com/NinjaParty/relay/core"
)

//go:embed icon.png
var iconData []byte

func runTray(relay *core.Relay, localUIAddr string, logger *zap.Logger) {
	systray.Run(func() {
		systray.SetIcon(iconData)
		systray.SetTitle("NinjaParty Relay")
		systray.SetTooltip("NinjaParty Relay")

		mTitle := systray.AddMenuItem("NinjaParty Relay", "")
		mTitle.Disable()

		mVersion := systray.AddMenuItem(fmt.Sprintf("Version: %s", version), "")
		mVersion.Disable()

		mStatus := systray.AddMenuItem("Status: Not in room", "")
		mStatus.Disable()

		systray.AddSeparator()
		mOpen := systray.AddMenuItem("Open UI", "Open the relay UI")
		mCopyCode := systray.AddMenuItem("Copy Room Code", "Copy the room code to clipboard")
		mCopyLink := systray.AddMenuItem("Copy Link", "Copy the room link to clipboard")
		mQuit := systray.AddMenuItem("Quit", "Close the relay")
		mQuit.SetTooltip("Quit the relay application")

		go func() {
			for {
				if relay.RoomCode() == "" {
					mStatus.SetTitle("Status: Not in room")
					mCopyCode.Hide()
					mCopyLink.Hide()
				} else {
					mStatus.SetTitle("Status: In room")
					mCopyCode.Show()
					mCopyLink.Show()
				}
				time.Sleep(5 * time.Second)
			}
		}()

		for {
			select {
			case <-mOpen.ClickedCh:
				browser.OpenURL(fmt.Sprintf("http://%s", localUIAddr))
			case <-mCopyCode.ClickedCh:
				if relay.RoomCode() != "" {
					clipboard.WriteAll(relay.RoomCode())
				}
			case <-mCopyLink.ClickedCh:
				if relay.RoomCode() != "" {
					clipboard.WriteAll(relay.FullLink())
				}
			case <-mQuit.ClickedCh:
				logger.Info("shutting down relay via tray")
				relay.Shutdown()
				os.Exit(0)
			}
		}
	}, func() {
		logger.Info("systray exited")
	})
}
