package main

import (
  "bufio"
  "encoding/json"
  "flag"
  "fmt"
  "os"
  "strings"
  "time"
)

type Event struct {
  Timestamp time.Time `json:"timestamp"`
  Op        string    `json:"op"`
  Path      string    `json:"path"`
}

func main() {
  journalPath := flag.String("journal", "journal.log", "Path to the journal file")
  threshold := flag.Int("threshold", 5, "Number of renames to trigger alert")
  window := flag.Int("window", 10, "Time window in seconds")
  flag.Parse()

  fmt.Printf("Monitoring %s for suspicious activity...\n", *journalPath)

  file, err := os.Open(*journalPath)
  if err != nil {
    panic(err)
  }
  defer file.Close()

  // Seek to end so we only process new entries
  file.Seek(0, os.SEEK_END)
  reader := bufio.NewReader(file)

  var eventTimes []time.Time

  for {
    line, err := reader.ReadString('\n')
    if err != nil {
      time.Sleep(500 * time.Millisecond)
      continue
    }

    line = strings.TrimSpace(line)
    if line == "" {
      continue
    }

    var ev Event
    if err := json.Unmarshal([]byte(line), &ev); err != nil {
      fmt.Println("Failed to parse:", line)
      continue
    }

    if ev.Op == "RENAME" || strings.Contains(ev.Path, ".locked") {
      eventTimes = append(eventTimes, ev.Timestamp)

      // Remove old entries outside window
      cutoff := time.Now().UTC().Add(-time.Duration(*window) * time.Second)
      var recent []time.Time
      for _, t := range eventTimes {
        if t.After(cutoff) {
          recent = append(recent, t)
        }
      }
      eventTimes = recent

      if len(eventTimes) >= *threshold {
        fmt.Printf("🚨 ALERT: %d suspicious renames in the last %d seconds at %s\n",
          len(eventTimes), *window, time.Now().Format(time.RFC3339))
        eventTimes = []time.Time{}
      }
    }
  }
}
