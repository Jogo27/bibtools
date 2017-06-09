package main

import "bufio"
import "fmt"
import "io"
import "os"
import "regexp"
import "time"

func ReadAuxFile(filename string, out chan<- string) {
  var file *os.File
  var err  error
  if file, err = os.Open(filename) ; err != nil {
    fmt.Fprintf(os.Stderr, "Unable to read %s\n", filename);
    close(out)
    return
  }

  scanner := bufio.NewScanner(file)
  rex := regexp.MustCompile("\\\\citation{([^\n}]+)}");
  sent := map[string] bool {}
  for scanner.Scan() {
    if match := rex.FindSubmatch(scanner.Bytes()) ; match != nil {
      ref := string(match[1])
      if !(sent[ref]) {
        sent[ref] = true
        out <- ref
      }
    }
  }

  fmt.Println("fini")
  close(out)
  file.Close()
}

type BibElement struct {
  name   string
  btype  string
  values map[string] string
}

func (self BibElement) Display (out io.Writer) {
  fmt.Fprintf(out, "@%s{%s,\n", self.name, self.btype)
  for key, val := range self.values {
    fmt.Fprintf(out, "\t%s = %s,\n", key, val)
  }
  fmt.Fprintln(out, "}\n")
}

func ReadBibFile (filename string, out chan<- BibElement) {
  var file *os.File
  var err  error
  if file, err = os.Open(filename) ; err != nil {
    fmt.Fprintf(os.Stderr, "Unable to read %s\n", filename);
    close(out)
    return
  }

  r_begin := regexp.MustCompile("^@([^\n{]+){([^\\s,]+),\\s*$");
  r_entry := regexp.MustCompile("^\\s*([^\\s=]+)\\s*=\\s*(.*?)(,)?\\s*$");
  r_cont  := regexp.MustCompile("^\\s*(.*?)(,)?\\s*$");
  r_end   := regexp.MustCompile("^}\\s*$");

  scanner := bufio.NewScanner(file)

  var elem BibElement
  var last string
  var results [][]byte
  const (
    s_outside = iota
    s_inside
    s_continue
  )
  state := s_outside
  for scanner.Scan() {
    switch state {

      case s_outside:
        if results = r_begin.FindSubmatch(scanner.Bytes()) ; results != nil {
          elem = BibElement{string(results[1]),
                            string(results[2]),
                            make(map[string]string)}
          state = s_inside
        }

      case s_inside:
        if results = r_entry.FindSubmatch(scanner.Bytes()) ; results != nil {
          key := string(results[1])
          elem.values[key] = string(results[2])
          if results[3] == nil {
            last = key
            state = s_continue
          }
        }

      case s_continue:
        if results = r_cont.FindSubmatch(scanner.Bytes()) ; results != nil {
          elem.values[last] += string(results[1])
          if results[2] != nil {
            state = s_inside
          }
        }
    } // swith
    if results = r_end.FindSubmatch(scanner.Bytes()) ; results != nil {
      out <- elem
      state = s_outside
    }
  } // for

  fmt.Println("fini")
  close(out)
  file.Close()
}

func main() {
  if len(os.Args) < 3 {
    fmt.Fprintf(os.Stderr, "Too few arguments\n");
    return;
  }

  chaux := make(chan string, 10)
  go ReadAuxFile(os.Args[1], chaux)

  for {
    line, ok := <-chaux
    if !ok { break }
    time.Sleep(time.Second/5)
    fmt.Println(line)
  }

  chbib := make(chan BibElement, 10)
  go ReadBibFile(os.Args[2], chbib)

  for {
    elem, ok := <-chbib
    if !ok { break }
    time.Sleep(time.Second/5)
    elem.Display(os.Stdout)
  }
}
