package notifier

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
)

// Notify triggers a system-level balloon notification on Windows.
func Notify(title, message string) {
	if runtime.GOOS != "windows" {
		return
	}

	// Escape single quotes for PowerShell compatibility
	safeTitle := strings.ReplaceAll(title, "'", "''")
	safeMsg := strings.ReplaceAll(message, "'", "''")

	script := fmt.Sprintf(
		"[void][System.Reflection.Assembly]::LoadWithPartialName('System.Windows.Forms'); "+
			"$obj = New-Object System.Windows.Forms.NotifyIcon; "+
			"$obj.Icon = [System.Drawing.SystemIcons]::Information; "+
			"$obj.BalloonTipTitle = '%s'; "+
			"$obj.BalloonTipText = '%s'; "+
			"$obj.Visible = $true; "+
			"$obj.ShowBalloonTip(5000);",
	safeTitle, safeMsg)

	cmd := exec.Command("powershell", "-NoProfile", "-Command", script)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	_ = cmd.Run()
}
