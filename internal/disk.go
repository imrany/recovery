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

    // Images
    "jpg":  {header: []byte{0xFF, 0xD8, 0xFF}, footer: []byte{0xFF, 0xD9}},
    "png":  {header: []byte{0x89, 0x50, 0x4E, 0x47}, footer: []byte{0x49, 0x45, 0x4E, 0x44, 0xAE, 0x42, 0x60, 0x82}},
    "gif":  {header: []byte{0x47, 0x49, 0x46, 0x38}, footer: []byte{0x3B}},
    "bmp":  {header: []byte{0x42, 0x4D}, footer: nil},
    "tiff": {header: []byte{0x49, 0x49, 0x2A, 0x00}, footer: nil},
    "webp": {header: []byte{0x52, 0x49, 0x46, 0x46}, footer: []byte{0x57, 0x45, 0x42, 0x50}},
    "svg":  {header: []byte{0x3C, 0x73, 0x76, 0x67}, footer: []byte{0x3C, 0x2F, 0x73, 0x76, 0x67}},
    "heic": {header: []byte{0x66, 0x74, 0x79, 0x70}, footer: nil}, // HEIC files often start with 'ftyp'
    "ico":  {header: []byte{0x00, 0x00, 0x01, 0x00}, footer: nil},
    "raw":  {header: []byte{0x49, 0x49, 0x2A, 0x00}, footer: nil}, // Common for RAW image formats
    "cr2":  {header: []byte{0x49, 0x49, 0x2A, 0x00}, footer: nil}, // Canon RAW files
    "nef":  {header: []byte{0x4E, 0x45, 0x46, 0x46}, footer: nil}, // Nikon RAW files
    "orf":  {header: []byte{0x4F, 0x52, 0x46, 0x00}, footer: nil}, // Olympus RAW files
    "arw":  {header: []byte{0x41, 0x52, 0x57, 0x00}, footer: nil}, // Sony RAW files
    "dng":  {header: []byte{0x44, 0x4E, 0x47, 0x00}, footer: nil}, // Adobe DNG files
    "psd":  {header: []byte{0x38, 0x42, 0x50, 0x53}, footer: nil}, // Photoshop files
    "heif": {header: []byte{0x66, 0x74, 0x79, 0x70}, footer: nil}, // HEIF files often start with 'ftyp'
    "avif": {header: []byte{0x66, 0x74, 0x79, 0x70}, footer: nil}, // AVIF files often start with 'ftyp'
    "jxl":  {header: []byte{0x0A, 0x4A, 0x58, 0x4C}, footer: nil}, // JPEG XL files
    "svgz": {header: []byte{0x1F, 0x8B, 0x08}, footer: []byte{
        0x1F, 0x8B, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xFF, 0x3C, 0xB2, 0xC1, 0x4A,0xC3, 0x30, 0x10, 0x85, 0xE1}}, // Compressed SVG files
    "x3d":  {header: []byte{0x3C, 0x78, 0x33, 0x64}, footer: []byte{0x3C, 0x2F, 0x78, 0x33, 0x64}},
    "hdr":  {header: []byte{0x23, 0x48, 0x44, 0x52}, footer: nil}, // High Dynamic Range images
    "exr":  {header: []byte{0x76, 0x2F, 0x31, 0x01}, footer: nil}, // OpenEXR files
    "xpm":  {header: []byte{0x2F, 0x2A, 0x20, 0x58}, footer: []byte{0x2A, 0x2F}}, // X PixMap files

    // Documents
    "pdf":  {header: []byte{0x25, 0x50, 0x44, 0x46}, footer: []byte{0x0A, 0x25, 0x25, 0x45, 0x4F, 0x46}},
    "docx": {header: []byte{0x50, 0x4B, 0x03, 0x04}, footer: nil}, 
    "xlsx": {header: []byte{0x50, 0x4B, 0x03, 0x04}, footer: nil}, 
    "pptx": {header: []byte{0x50, 0x4B, 0x03, 0x04}, footer: nil}, 
    "txt":  {header: []byte{0xEF, 0xBB, 0xBF}, footer: nil}, 
    "rtf":  {header: []byte{0x7B, 0x5C, 0x72, 0x74, 0x66}, footer: []byte{0x7D}}, 
    "epub": {header: []byte{0x50, 0x4B, 0x03, 0x04}, footer: nil},

    // Videos
    "mp4":  {header: []byte{0x66, 0x74, 0x79, 0x70}, footer: nil},
    "avi":  {header: []byte{0x52, 0x49, 0x46, 0x46}, footer: nil}, 
    "mkv":  {header: []byte{0x1A, 0x45, 0xDF, 0xA3}, footer: nil}, 
    "mov":  {header: []byte{0x6D, 0x6F, 0x6F, 0x76}, footer: nil}, 
    "wmv":  {header: []byte{0x30, 0x26, 0xB2, 0x75}, footer: nil}, 
    "flv":  {header: []byte{0x46, 0x4C, 0x56}, footer: nil},

    // Audio
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

// Scan disk sectors for specified file types
func Scan(diskPath string, selectedType string) {
    if selectedType != "" { 
        fmt.Println("[*] Scanning disk sectors for deleted files...")
        fmt.Println("[*] Starting recovery for file type:", selectedType)
    
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
        sector := 0
    
        for {
            n, err := file.Read(buffer)
            if n == 0 || err != nil {
                break
            }
    
            // Check if the sector contains any known file signature
            for fileType, sig := range fileSignatures {
                if selectedType != "all" && fileType != selectedType {
                    continue // Skip files that don’t match the selected type
                }
    
                if bytes.Equal(buffer[:len(sig.header)], sig.header) {
                    fmt.Printf("[+] Found %s header, starting recovery...\n", fileType)
                    foundHeader = true
                    detectedFileType = fileType
                }
            }
    
            if foundHeader {
                recoveredData = append(recoveredData, buffer[:n]...)
    
                if sig, exists := fileSignatures[detectedFileType]; exists && sig.footer != nil && bytes.Contains(buffer, sig.footer) {
                    fmt.Printf("[+] Found %s footer, saving file...\n", detectedFileType)
                    recoverFile(recoveredData, fmt.Sprintf("recovered_%d.%s", sector, detectedFileType))
                    recoveredData = nil
                    foundHeader = false
                } else if sig.footer == nil && len(recoveredData) > 1000000 { // Recover files without footers after ~1MB
                    fmt.Printf("[+] Recovering %s file...\n", detectedFileType)
                    recoverFile(recoveredData, fmt.Sprintf("recovered_%d.%s", sector, detectedFileType))
                    recoveredData = nil
                    foundHeader = false
                }
            }

            sector ++
        }
    }
}

// Save the fully recovered file
func recoverFile(data []byte, outputPath string) {
    destDir := "./recovered/"

    if _, err := os.Stat(destDir); os.IsNotExist(err) {
        os.Mkdir(destDir, 0755)
    }

    file, err := os.OpenFile(destDir + outputPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
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
    if filePath !=""{
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
}

// Function to list partitions
func ListPartitions(partitions *bool, diskPath string) {
    if *partitions&&diskPath != "" {
        fmt.Println("[*] Listing partitions on disk:", diskPath)
        if _, err := os.Stat(diskPath); os.IsNotExist(err) {
            fmt.Println("[-] Disk not found:", diskPath)
            return
        }
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
}