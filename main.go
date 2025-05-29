package main

import (
	// "fyne.io/fyne/v2/app"
	// "fyne.io/fyne/v2/container"
	// "fyne.io/fyne/v2/widget"
	"fmt"
	"os"
	"runtime"

	disk "github.com/imrany/recovery/internals"
)

// GUI for recovery tool
func main() {
    // recoveryApp := app.New()
    // window := recoveryApp.NewWindow("File Recovery Tool")

    // scanButton := widget.NewButton("Scan for Deleted Files", func() {
    //     fmt.Println("Scanning disk sectors...")
    // })

    // recoverButton := widget.NewButton("Recover Selected Files", func() {
    //     fmt.Println("Recovering files...")
    // })

    // window.SetContent(container.NewVBox(
    //     widget.NewLabel("Welcome to File Recovery"),
    //     scanButton,
    //     recoverButton,
    // ))

    // window.ShowAndRun()

	var diskPath string
	switch {
		case os.Getenv("WSL_DISTRO_NAME") != "":
			diskPath = "/mnt/c"
		case os.Getenv("OS") == "Windows_NT":
			diskPath = "C:\\"
		case os.Getenv("XDG_SESSION_TYPE") != "":
			diskPath = "/dev/sda"
		case runtime.GOOS == "darwin":
			diskPath = "/dev/disk0"
		default:
			diskPath = "/dev/sda"
	}
	fmt.Println("Scanning disk sectors for deleted files...")
	disk.Scan(diskPath)
}