package main

import (
    "fmt"
    "os"
    "path/filepath"
)

// Define file signatures (for basic recovery)
var fileSignatures = map[string][]byte{
    "jpg": {0xFF, 0xD8, 0xFF},
    "png": {0x89, 0x50, 0x4E, 0x47},
    "pdf": {0x25, 0x50, 0x44, 0x46},
}

// Scan for deleted files
func scanDeletedFiles(directory string) {
    err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        // Check for hidden/deleted files
        if info.Size() > 0 && isRecoverable(path) {
            fmt.Println("[+] Recoverable file found:", path)
            recoverFile(path)
        }
        return nil
    })
    
    if err != nil {
        fmt.Println("Error scanning:", err)
    }
}

// Check if file matches known signatures
func isRecoverable(filePath string) bool {
    file, err := os.Open(filePath)
    if err != nil {
        return false
    }
    defer file.Close()

    header := make([]byte, 4)
    _, err = file.Read(header)
    if err != nil {
        return false
    }

    for _, sig := range fileSignatures {
        if len(header) >= len(sig) && compareBytes(header[:len(sig)], sig) {
            return true
        }
    }
    return false
}

// Compare file signature bytes
func compareBytes(a, b []byte) bool {
    for i := range b {
        if a[i] != b[i] {
            return false
        }
    }
    return true
}

// Recover the file
func recoverFile(filePath string) {
	err:=os.Mkdir("./recovered", 0755)
	if err != nil && !os.IsExist(err) {
		fmt.Println("[-] Error creating recovery directory:", err)
		return
	}

    destPath := "./recovered/" + filepath.Base(filePath)
    err = os.Rename(filePath, destPath)
    if err != nil {
        fmt.Println("[-] Recovery failed:", err)
    } else {
        fmt.Println("[+] Recovered:", destPath)
    }
}

func main() {
    scanDir := "./disk_image" // Set the directory to scan
    fmt.Println("Scanning for deleted files...")
    scanDeletedFiles(scanDir)
}
