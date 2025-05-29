package disk

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"syscall"
)

// File signatures for common types
var fileSignatures = map[string]struct {
    header []byte
    footer []byte
}{
    "pdf":  {header: []byte{0x25, 0x50, 0x44, 0x46}, footer: []byte{0x0A, 0x25, 0x25, 0x45, 0x4F, 0x46}}, // PDF ends with %%EOF
    "docx": {header: []byte{0x50, 0x4B, 0x03, 0x04}, footer: nil}, // DOCX shares ZIP structure
    "xlsx": {header: []byte{0x50, 0x4B, 0x03, 0x04}, footer: nil},
    "pptx": {header: []byte{0x50, 0x4B, 0x03, 0x04}, footer: nil},
    "txt":  {header: []byte{0xEF, 0xBB, 0xBF}, footer: nil}, // UTF-8 BOM header (optional)
    "rtf":  {header: []byte{0x7B, 0x5C, 0x72, 0x74, 0x66}, footer: []byte{0x7D}}, // RTF starts with "{\rtf"
    "epub": {header: []byte{0x50, 0x4B, 0x03, 0x04}, footer: nil}, // EPUB is ZIP-based

    // Video formats
    "mp4":  {header: []byte{0x66, 0x74, 0x79, 0x70}, footer: nil},
    "avi":  {header: []byte{0x52, 0x49, 0x46, 0x46}, footer: nil}, // RIFF format
    "mkv":  {header: []byte{0x1A, 0x45, 0xDF, 0xA3}, footer: nil}, // Matroska format
    "mov":  {header: []byte{0x6D, 0x6F, 0x6F, 0x76}, footer: nil},
    "wmv":  {header: []byte{0x30, 0x26, 0xB2, 0x75}, footer: nil},
    "flv":  {header: []byte{0x46, 0x4C, 0x56}, footer: nil},

    // Audio formats
    "mp3":  {header: []byte{0x49, 0x44, 0x33}, footer: nil},
    "wav":  {header: []byte{0x52, 0x49, 0x46, 0x46}, footer: []byte{0x57, 0x41, 0x56, 0x45}},
    "aac":  {header: []byte{0xFF, 0xF1}, footer: nil},
    "flac": {header: []byte{0x66, 0x4C, 0x61, 0x43}, footer: nil},
    "ogg":  {header: []byte{0x4F, 0x67, 0x67, 0x53}, footer: nil},
    "m4a":  {header: []byte{0x00, 0x00, 0x00, 0x18, 0x66, 0x74, 0x79, 0x70}, footer: nil},

    // Archives
    "zip":  {header: []byte{0x50, 0x4B, 0x03, 0x04}, footer: nil},
    "rar":  {header: []byte{0x52, 0x61, 0x72, 0x21, 0x1A, 0x07}, footer: nil},
    "7z":   {header: []byte{0x37, 0x7A, 0xBC, 0xAF, 0x27, 0x1C}, footer: nil},
    "tar":  {header: []byte{0x75, 0x73, 0x74, 0x61, 0x72}, footer: nil},
    "iso":  {header: []byte{0x43, 0x44, 0x30, 0x30, 0x31}, footer: nil},
    "dmg":  {header: []byte{0x78, 0x01, 0x73, 0x0D, 0x62, 0x62}, footer: nil},
}

// Scan disk sectors for multiple file types
func Scan(diskPath string) {
    file, err := os.Open(diskPath)
    if err != nil {
        fmt.Println("[-] Failed to open disk:", err)
        return
    }
    defer file.Close()

    buffer := make([]byte, 512)
    var recoveredData []byte
    var foundHeader bool
    var detectedFileType string

    for {
        n, err := file.Read(buffer)
        if n == 0 || err != nil {
            break
        }

        // Check if the sector contains any known file signature
        for fileType, sig := range fileSignatures {
            if bytes.Equal(buffer[:len(sig.header)], sig.header) {
                fmt.Printf("[+] Found %s header, starting recovery...\n", fileType)
                foundHeader = true
                detectedFileType = fileType
            }
        }

        if foundHeader {
            recoveredData = append(recoveredData, buffer[:n]...)

            // If the file type has a known footer, check if it's fully recovered
            if sig, exists := fileSignatures[detectedFileType]; exists && sig.footer != nil && bytes.Contains(buffer, sig.footer) {
                fmt.Printf("[+] Found %s footer, saving file...\n", detectedFileType)
                recoverFile(recoveredData, fmt.Sprintf("recovered_%s.%s", detectedFileType, detectedFileType))
                recoveredData = nil
                foundHeader = false
            } else if sig.footer == nil && len(recoveredData) > 1000000 { // Recover files without footers after ~1MB
                fmt.Printf("[+] Recovering %s file...\n", detectedFileType)
                recoverFile(recoveredData, fmt.Sprintf("recovered_%s.%s", detectedFileType, detectedFileType))
                recoveredData = nil
                foundHeader = false
            }
        }
    }
}

// Save the fully recovered file
func recoverFile(data []byte, outputPath string) {
    destDir := "./recovered/"

    // Ensure recovery folder exists
    if _, err := os.Stat(destDir); os.IsNotExist(err) {
        os.Mkdir(destDir, 0755)
    }

    // Write recovered file
    file, err := os.Create(destDir + outputPath)
    if err != nil {
        fmt.Println("[-] Failed to save recovered file:", err)
        return
    }
    defer file.Close()

    _, err = file.Write(data)
    if err != nil {
        fmt.Println("[-] Error writing recovered file:", err)
    } else {
        fmt.Println("[+] Successfully recovered:", outputPath)
    }
}

// Recover metadata (timestamps, original path)
func GetFileMetadata(filePath string) {
    fileInfo, err := os.Stat(filePath)
    if err != nil {
        fmt.Println("[-] Error accessing file:", err)
        return
    }

    stat := fileInfo.Sys().(*syscall.Stat_t)
    fmt.Println("[+] Metadata for:", filePath)
    fmt.Println(" - Size:", fileInfo.Size())
    fmt.Println(" - Last Modified:", fileInfo.ModTime())
    fmt.Println(" - Inode:", stat.Ino)
}

// Function to list partitions
func ListPartitions() {
    fmt.Println("Scanning for lost partitions...")
    var cmd string
    var args []string
    switch {
        case os.Getenv("OS") == "Windows_NT":
            cmd = "wmic"
            args = []string{"logicaldisk", "get", "DeviceID,VolumeName,FileSystem,Size,FreeSpace"}
        case runtime.GOOS == "darwin":
            cmd = "diskutil"
            args = []string{"list"}
        default:
            cmd = "lsblk"
            args = []string{"-o", "NAME,SIZE,FSTYPE,MOUNTPOINT"}
    }
    
    execute := exec.Command(cmd, args...)
    output, err := execute.Output()
    if err != nil {
        fmt.Println("[-] Error scanning partitions:", err)
        return
    }

    fmt.Println("[+] Detected partitions:\n", string(output))
}