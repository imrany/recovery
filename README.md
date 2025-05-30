This tool scans disk sectors for deleted files and attempts to recover them based on file signatures.


You can specify the type of files to recover using the -type flag.

---

### **🚀 Usage**
1. To scan and recover specific files `recovery -disk=<diskpath> -type=file_extension` 
2. To scan and recover all files `recovery -disk=<diskpath> -type=all`
3. List partitions on a disk `recovery -disk=<diskpath> -partitions`
4. Check a files metadata `recovery -info=file_path` 

---

### **🔥 Get your recovered files**
1️⃣ Recovered files would be saved as **`./recovered/recovered_<sector>.<ext>`**  
