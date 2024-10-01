package main

import (
	"flag"
	"fmt"
	"log"
	"io"
	"os"
	"path/filepath"
)

type config struct {
  // extension to filter out
  ext string
  // min file size
  size int64
  // lsit files
  list bool
  del bool
  // log destination writer
  wLog io.Writer
  // archive directory
  archive string
}

func main() {
  // Parsing command line flags
  root := flag.String("root", ".", "Root directory to start")
  // Action options
  list := flag.Bool("list", false, "List files only")
  del := flag.Bool("del", false, "Delete files")
  // Filter options
  ext := flag.String("ext", "", "File extension to filter out")
  size := flag.Int64("size", 0, "Minimum file size")
  logFile := flag.String("log", "", "Log deletes to this file")
  archive := flag.String("archive", "", "Archive directory")
  flag.Parse()

  var (
    f = os.Stdout
    err error
  )

  if *logFile != "" {
    f, err = os.OpenFile(*logFile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
    if err != nil {
      fmt.Fprintln(os.Stderr, err)
      os.Exit(1)
    }
    defer f.Close()
  }

  c := config{
    ext: *ext,
    size: *size,
    list: *list,
    del: *del,
    wLog: f,
    archive: *archive,
  }

  if err := run(*root, os.Stdout, c); err != nil {
    fmt.Fprintln(os.Stderr, err)
    os.Exit(1)
  }
}

func run(root string, out io.Writer, cfg config) error {
  delLogger := log.New(cfg.wLog, "DELETED FILE: ", log.LstdFlags)
  return filepath.Walk(root,
  func(path string, info os.FileInfo, err error) error {
    if err != nil {
      return err
    }

    if filterOut(path, cfg.ext, cfg.size, info) {
      return nil
    }

    // If list was explicitly set, don't do anything else
    if cfg.list {
      return listFile(path, out)
    }

    if cfg.del {
      return delFile(path, delLogger)
    }

    if cfg.archive != "" {
      if err := archiveFile(cfg.archive, root, path); err != nil {
        return err
      }
    }

    // List is the default option if nothing else was set
    return listFile(path, out)
  })
}
