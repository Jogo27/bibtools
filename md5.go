package main

import "fmt"
import "io"
import md5 "crypto/md5"
import "os"
import "regexp"

const nbworkers = 2

type FileInfo struct {
  name string
  md5  string
}

func hashMd5(filename string) (hsum string, err error) {

  var file *os.File
  file, err = os.Open(filename)
  if err != nil { return }
  defer file.Close()

  h := md5.New()
  _, err = io.Copy(h, file)
  if err != nil { return }

  hsum = fmt.Sprintf("%x",h.Sum(nil))
  return
}

func FileInfoRoutine(in <-chan string, out chan<- interface{}) {
  for {
    filename, ok := <-in
    if !ok { break }

    hsum, err := hashMd5(filename)
    if err != nil { continue }
    out <- FileInfo{filename, hsum}
  }
  out <- nil
}

func main() {
  if len(os.Args) < 2 {
    fmt.Fprintf(os.Stderr, "Too few arguments\n");
    return;
  }

  chfn := make(chan string, 32)
  go ScanDirMatch(os.Args[1], regexp.MustCompile("\\.pdf$"), chfn)

  chfi := make(chan interface{}, 32)

  for i := 0; i < nbworkers ; i+= 1 {
    go FileInfoRoutine(chfn, chfi)
  }

  for nb := nbworkers; nb > 0; {
    poly, ok := <-chfi
    if !ok { break }
    switch elem := poly.(type) {
      case FileInfo:
        fmt.Printf("%-70s %s\n", elem.name, elem.md5)
      case nil:
        nb -= 1
      default:
        panic("Impossible type received")
    }
  }
  close(chfi)
}
