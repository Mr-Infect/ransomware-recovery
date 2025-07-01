package main

import (
  "bufio"
  "encoding/json"
  "flag"
  "fmt"
  "io/ioutil"
  "os"
  "path/filepath"
  "strings"
  "time"
)

type Event struct {
  Timestamp time.Time `json:"timestamp"`
  Op        string    `json:"op"`
  Path      string    `json:"path"`
}

func main() {
  sinceStr := flag.String("since", "", "Restore files affected after this timestamp (RFC3339)")
  journalPath := flag.String("journal", "journal.log", "Path to journal file")
  backupDir := flag.String("backup", ".backup", "Backup folder path")
  targetDir := flag.String("target", "", "Original watched directory path")
  flag.Parse()

  if *sinceStr == "" || *targetDir == "" {
    fmt.Println("Usage: rollback --since <timestamp> --target <watched-folder>")
    return
  }

  sinceTime, err := time.Parse(time.RFC3339, *sinceStr)
  if err != nil {
    fmt.Println("Invalid time:", err)
    return
  }

  journal, err := os.Open(*journalPath)
  if err != nil {
    fmt.Println("Error opening journal:", err)
    return
  }
  defer journal.Close()

  scanner := bufio.NewScanner(journal)
  restored := 0

  for scanner.Scan() {
    line := scanner.Text()
    var ev Event
    if err := json.Unmarshal([]byte(line), &ev); err != nil {
      continue
    }

    if ev.Timestamp.Before(sinceTime) {
      continue
    }

    // Restore on RENAME or .locked file
    if ev.Op == "RENAME" || strings.Contains(ev.Path, ".locked") {
      // original path (before encryption)
      origRel := strings.TrimSuffix(ev.Path, ".locked")
      backupPath := filepath.Join(*backupDir, origRel)
      restorePath := filepath.Join(*targetDir, origRel)

      if _, err := os.Stat(backupPath); err == nil {
        os.MkdirAll(filepath.Dir(restorePath), 0755)
        data, err := ioutil.ReadFile(backupPath)
        if err != nil {
          fmt.Println("Failed to read backup:", backupPath)
          continue
        }
        err = ioutil.WriteFile(restorePath, data, 0644)
        if err != nil {
          fmt.Println("Failed to restore:", restorePath)
          continue
        }
        fmt.Printf("✅ Restored: %s\n", restorePath)
        restored++
      }
    }
  }

  if restored == 0 {
    fmt.Println("⚠️ No files restored.")
  } else {
    fmt.Printf("🎉 Done. %d files restored.\n", restored)
  }
}
