//go:build windows

package utils

import (
	"log"
	"syscall"
	"unsafe"
)

var (
	kernel32         = syscall.NewLazyDLL("kernel32.dll")
	user32           = syscall.NewLazyDLL("user32.dll")
	getConsoleWindow = kernel32.NewProc("GetConsoleWindow")
	setConsoleTitleW = kernel32.NewProc("SetConsoleTitleW")
	showWindow       = user32.NewProc("ShowWindow")
	getSystemMenu    = user32.NewProc("GetSystemMenu")
	deleteMenu       = user32.NewProc("DeleteMenu")
)

const (
	SW_HIDE            = 0
	SW_MINIMIZE        = 6
	SW_SHOWMINNOACTIVE = 7
	SC_CLOSE           = 0xF060
	MF_BYCOMMAND       = 0x00000000
)

// SetupServerConsole sets the console window title, minimizes it to the taskbar,
// and disables the 'X' close button to protect against accidental closure.
func SetupServerConsole(title string, minimize bool, disableCloseButton bool) {
	hwnd, _, _ := getConsoleWindow.Call()
	if hwnd == 0 {
		return
	}

	if title != "" {
		pTitle, err := syscall.UTF16PtrFromString(title)
		if err == nil {
			_, _, _ = setConsoleTitleW.Call(uintptr(unsafe.Pointer(pTitle)))
		}
	}

	if disableCloseButton {
		// Disable the 'X' close button on the console window menu
		hMenu, _, _ := getSystemMenu.Call(hwnd, 0)
		if hMenu != 0 {
			_, _, _ = deleteMenu.Call(hMenu, SC_CLOSE, MF_BYCOMMAND)
			log.Printf("[Server] Schließen-Button (X) der Konsole zum Schutz vor versehentlichem Beenden deaktiviert.")
		}
	}

	if minimize {
		// Minimize console window to the taskbar
		_, _, _ = showWindow.Call(hwnd, SW_MINIMIZE)
		log.Printf("[Server] Serverfenster wurde minimiert in die Taskleiste abgelegt.")
	}
}
