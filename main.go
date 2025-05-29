package main

import (
	// "fyne.io/fyne/v2/app"
	// "fyne.io/fyne/v2/container"
	// "fyne.io/fyne/v2/widget"
	"flag"
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

	var fileType string
	var diskPath string
	var fileInfo string

	flag.StringVar(&fileType, "type", "", "Specify file type to recover (e.g., 'images', 'video', 'pdf', 'all', 'audio', 'documents', 'archives', 'executable', 'text', 'code', 'fonts', 'database', 'email', 'backup'). Default is 'all'.")
	flag.StringVar(&fileInfo, "info", "", "Display information about the specified file")
	flag.StringVar(&diskPath, "disk", "", "Specify the disk path (e.g., /dev/sda, C:\\, /dev/disk0)")
	partitions := flag.Bool("partitions", false, "List partitions")
	flag.Parse()

	if diskPath == "" {
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
	}
	disk.ListPartitions(partitions)
	disk.GetFileMetadata(fileInfo)
	disk.Scan(diskPath, fileType)
}