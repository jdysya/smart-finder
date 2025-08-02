package utils

import (
	"fmt"
	"os/exec"
	"runtime"
)

func RevealInExplorer(filePath string) error {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("explorer", "/select,", filePath).Start()
	case "darwin":
		return exec.Command("open", "-R", filePath).Start()
	case "linux":
		// 以 Nautilus 为例，其他桌面环境可自行扩展
		return exec.Command("nautilus", "--select", filePath).Start()
	default:
		return nil
	}
}

func OpenURL(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		return fmt.Errorf("unsupported platform")
	}
	return cmd.Start()
}
