package tray

import (
	"fmt"
	"log"

	"smart-finder/client/internal/icon"
	"smart-finder/client/internal/utils"

	"github.com/getlantern/systray"
)

var (
	mStatus *systray.MenuItem
)

const (
	controlPanelURL = "http://127.0.0.1:8964"
)

func Run(onReady func(), onExit func()) {
	systray.Run(func() {
		onReady()
		onReadyWrapper()
	}, func() {
		onExit()
		onExitWrapper()
	})
}

func onReadyWrapper() {
	systray.SetIcon(icon.Data)
	systray.SetTitle("Smart Finder")
	systray.SetTooltip("Smart Finder is running")

	mOpen := systray.AddMenuItem("Open Control Panel", "Open the web UI")
	systray.AddSeparator()
	mStatus = systray.AddMenuItem("Status: Initializing...", "Current status")
	mStatus.Disable()
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quit the whole app")

	go func() {
		for {
			select {
			case <-mOpen.ClickedCh:
				err := utils.OpenURL(controlPanelURL)
				if err != nil {
					log.Printf("Failed to open control panel: %v", err)
				}
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

func onExitWrapper() {
	log.Println("Exiting tray...")
}

func UpdateStatus(status string) {
	if mStatus != nil {
		mStatus.SetTitle(fmt.Sprintf("Status: %s", status))
	}
}
