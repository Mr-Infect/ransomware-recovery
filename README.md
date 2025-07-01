
### 🛡️ Ransomware Recovery MVP

A minimalistic, working proof-of-concept for detecting and recovering from ransomware attacks on live file systems. Built entirely in Go, this MVP includes:

- ✅ Real-time file system monitoring  
- 🚨 Rule-based ransomware detection  
- ♻️ One-click rollback of encrypted files  

> This is a cybersecurity master’s project focused on simplicity, transparency, and extensibility.

---

## 📦 Components

| Tool        | Description                                           |
|-------------|-------------------------------------------------------|
| `agent.go`  | Watches a target folder and logs file events         |
| `detector.go` | Detects suspicious file rename activity (e.g., `.locked`) |
| `rollback.go` | Restores affected files from `.backup/`            |

---

## 🚀 Getting Started

### ✅ 1. Clone the repo
```bash
git clone https://github.com/Mr-Infect/ransomware-recovery.git
cd ransomware-recovery
````

### ✅ 2. Build the binaries

```bash
go build -o agent agent.go
go build -o detector detector.go
go build -o rollback rollback.go
```

### ✅ 3. Run the agent

```bash
./agent --watch ./demo-folder
```

This will:

* Watch `./demo-folder` for changes
* Log events to `journal.log`
* Backup original files into `.backup/`

### ✅ 4. Simulate a ransomware attack

```bash
mv demo-folder/file1.txt demo-folder/file1.txt.locked
mv demo-folder/file2.txt demo-folder/file2.txt.locked
```

### ✅ 5. Run the detector

```bash
./detector --journal journal.log --threshold 2 --window 10
```

If threshold is breached, you’ll see:

```
🚨 ALERT: 2 suspicious renames in the last 10 seconds
```

### ✅ 6. Roll back affected files

```bash
./rollback --since 2025-07-01T14:23:10Z --target ./demo-folder
```

> Restores only files modified after the given timestamp from `.backup/`.

---

## 🧱 Folder Structure

```
.
├── agent.go          # File monitoring and journaling
├── detector.go       # Ransomware detection
├── rollback.go       # File recovery engine
├── journal.log       # Logs of file events
├── .backup/          # Auto-saved pre-attack file copies
└── demo-folder/      # The folder under attack (simulated)
```

---

## 🔮 Future Roadmap

* Configurable rules via `config.json`
* Auto-rollback trigger after detection
* CLI testing simulator for ransomware events
* Incident report generator (Markdown/JSON)
* System hardening + service mode

---

## ⚖️ License

This project is licensed under the MIT License.
Feel free to fork, modify, and build on it.

---

## 👨‍💻 Author

**Mr-Infect**
Cybersecurity & AI Engineer
[github.com/Mr-Infect](https://github.com/Mr-Infect)

```
