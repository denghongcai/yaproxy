package main

import (
	"github.com/denghongcai/yaproxy/icon"
	"github.com/getlantern/systray"
	"os"
	"syscall"
)

const (
	ATTACH_PARENT_PROCESS = ^uint32(0) // (DWORD)-1
)

var (
	modkernel32 = syscall.NewLazyDLL("kernel32.dll")

	procAttachConsole = modkernel32.NewProc("AttachConsole")
)

func AttachConsole(dwParentProcess uint32) (ok bool) {
	r0, _, _ := syscall.Syscall(procAttachConsole.Addr(), 1, uintptr(dwParentProcess), 0, 0)
	ok = bool(r0 != 0)
	return
}

func onReady() {
	systray.SetIcon(icon.Data)
	systray.SetTitle("yaproxy")
	systray.SetTooltip("(๑•́ ₃ •̀๑)")
	mQuit := systray.AddMenuItem("Quit", "Quit yaproxy")
	go func() {
		for {
			select {
			case <-mQuit.ClickedCh:
				systray.Quit()
				os.Exit(0)
				return
			}
		}
	}()
}

func main() {
	systray.Run(onReady)
	AttachConsole(ATTACH_PARENT_PROCESS)
	App()
}
