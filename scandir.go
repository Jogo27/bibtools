package main

import "os"
import "path"
import "regexp"

type StringFifo struct {
  slice []string
  nextRead  int
  nextWrite int
  begin int
  end   int
}

func CreateStringFifo(initsize int) *StringFifo {
  ret := new(StringFifo)
  ret.slice = make([]string, 0, initsize)
  ret.nextWrite = 0
  ret.nextRead  = 0
  ret.begin = -1
  ret.end   = -1
  return ret
}

func (this *StringFifo) IsEmpty() bool {
  return this.nextRead == this.nextWrite
}

func (this *StringFifo) Pop() (out string, ok bool) {
  if this.IsEmpty() {
    return "", false
  }

  ok  = true
  out = this.slice[this.nextRead]

  this.nextRead += 1
  if this.nextRead == this.end {
    this.nextRead = 0
    if this.begin == -1 { this.end = -1 }
  } else if this.nextRead == this.begin {
    this.nextRead = this.end
    this.begin = -1
    this.end   = -1
  }
  return
}

func (this *StringFifo) Push(str string) {
  if this.IsEmpty() {
    this.nextRead  = 0
    this.nextWrite = 0
    this.begin = -1
    this.end   = -1
  }
  if this.nextWrite >= len(this.slice) {
    if this.nextWrite >= cap(this.slice) && this.begin == -1 && this.end == -1 && this.nextRead > 0 {
      this.end   = this.nextWrite
      this.nextWrite = 0
      // we do NOT return here
    } else {
      this.slice = append(this.slice, str)
      this.nextWrite += 1
      return
    }
  }
  this.slice[this.nextWrite] = str
  this.nextWrite += 1
  if this.nextWrite == this.nextRead {
    this.begin = this.nextRead
    this.nextWrite = this.end
  }
}

func ScanDirMatch(basedir string, pattern *regexp.Regexp, out chan<- string) {
  os.Chdir(basedir)

  todo := CreateStringFifo(28)
  todo.Push(".")

  for !todo.IsEmpty() {

    filename, _ := todo.Pop()
    file, err := os.Open(filename)
    if err != nil { continue }

    list, err := file.Readdir(-1)
    file.Close()
    if err != nil { continue }
    for _, entry := range list {
      fullname := path.Join(filename, entry.Name())
      switch {
        case entry.IsDir():
          todo.Push(fullname)
        case pattern.MatchString(entry.Name()):
          out <- fullname
      }
    }

  }

  close(out)
}

/*
*/
