package main

import (
  "encoding/json"
  "flag"
  "fmt"
  "io"
  "io/ioutil"
  "log"
  "os"
  "path/filepath"
  "sync"
  "time"

  "github.com/fsnotify/fsnotify"
)

// Event represents one file-system change
type Event struct {
  Timestamp time.Time `json:"timestamp"`
  Op        string    `json:"op"`
  Path      string    `json:"path"`
}

func main() {
  // 1. Parse flags
  watchDir := flag.String("watch", "", "Directory to watch")
  journalFile := flag.String("journal", "journal.log", "Path to journal file")
  backupDir := flag.String("backup", ".backup", "Directory for backups")
  flag.Parse()

  if *watchDir == "" {
    log.Fatal("You must specify --watch")
  }

  // 2. Prepare backup dir
  os.MkdirAll(*backupDir, 0755)

  // 3. Open journal for appending
  jf, err := os.OpenFile(*journalFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
  if err != nil {
    log.Fatalf("Failed to open journal: %v", err)
  }
  defer jf.Close()
  mutex := &sync.Mutex{} // protect journal writes

  // 4. Create watcher
  watcher, err := fsnotify.NewWatcher()
  if err != nil {
    log.Fatal(err)
  }
  defer watcher.Close()

  // 5. Walk the tree to add watches
  filepath.Walk(*watchDir, func(path string, info os.FileInfo, err error) error {
    if info.IsDir() {
      watcher.Add(path)
    }
    return nil
  })

  // 6. Event loop
  log.Printf("Watching %s …\n", *watchDir)
  for {
    select {
    case ev, ok := <-watcher.Events:
      if !ok {
        return
      }
      // 6a. On Create/Write/Rename/Delete → record and backup
      if ev.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Rename|fsnotify.Remove) != 0 {
        go handleEvent(ev, *watchDir, *backupDir, jf, mutex)
      }
      // 6b. If new dir created, add watch
      if ev.Op&fsnotify.Create != 0 {
        info, err := os.Stat(ev.Name)
        if err == nil && info.IsDir() {
          watcher.Add(ev.Name)
        }
      }

    case err, ok := <-watcher.Errors:
      if !ok {
        return
      }
      log.Println("Watcher error:", err)
    }
  }
}

func handleEvent(ev fsnotify.Event, root, backupDir string, jf io.Writer, m *sync.Mutex) {
  rel, _ := filepath.Rel(root, ev.Name)
  // 1. On first Write/Rename of a file, back it up
  if ev.Op&(fsnotify.Write|fsnotify.Rename) != 0 {
    backupPath := filepath.Join(backupDir, rel)
    os.MkdirAll(filepath.Dir(backupPath), 0755)
    if _, err := os.Stat(backupPath); os.IsNotExist(err) {
      // copy original
      data, err := ioutil.ReadFile(ev.Name)
      if err == nil {
        ioutil.WriteFile(backupPath, data, 0644)
      }
    }
  }
  // 2. Log the event
  e := Event{time.Now().UTC(), ev.Op.String(), rel}
  blob, _ := json.Marshal(e)
  m.Lock()
  jf.Write(blob)
  jf.Write([]byte("\n"))
  m.Unlock()
  fmt.Printf("EVENT: %s %s\n", e.Op, e.Path)
}
