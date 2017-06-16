package main

import "bufio"
import "fmt"
import "io"
import "os"
import "regexp"
import "unicode"
import utf8 "unicode/utf8"

var display_order = []string{"author", "title", "journal", "booktitle", "year"}

type BibElement struct {
  name   string
  btype  string
  values map[string] string
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
          elem = BibElement{string(results[2]),
                            string(results[1]),
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
          elem.values[last] += " " + string(results[1])
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

  close(out)
  file.Close()
}

var display_order_set = func(tab []string) map[string] bool {
  ret := make(map[string] bool)
  for _, key := range tab {
    ret[key] = true
  }
  return ret
}(display_order)

func WrapLineSlice(length int, input string) [][]byte {
  inputlen := len(input)
  ret := make([][]byte, 0, (inputlen / length) + 1)
  b := []byte(input)

  begin := 0
  last := -1 // last space
  last_dep := 0
  for i := 0; i < inputlen; {
    r, dep := utf8.DecodeRune(b[i:]) // perhaps it's slow
    if unicode.IsSpace(r) {
      last = i
      last_dep = dep
    }
    if i + dep - begin > length {
      if last > 0 {
        ret = append(ret, b[begin:last])
        begin = last + last_dep
      } else {
        ret = append(ret, b[begin:i])
        begin = i
      }
      last = -1
    }
    i += dep
  }

  return append(ret, b[begin:inputlen])
}

func print_bib_attribute(out io.Writer, key string, val string) {
  lines := WrapLineSlice(80, val)
  fmt.Fprintf(out, "  %-12s = %s", key, lines[0])
  for _, line := range lines[1:] {
    fmt.Fprintf(out, "\n%18s%s", "", line)
  }
  fmt.Fprintf(out, ",\n")
}

func (self BibElement) Display (out io.Writer) {
  fmt.Fprintf(out, "@%s{%s,\n", self.btype, self.name)
  for _, key := range display_order {
    if self.values[key] != "" {
      print_bib_attribute(out, key, self.values[key])
    }
  }
  for key, val := range self.values {
    if ! display_order_set[key] {
      print_bib_attribute(out, key, val)
    }
  }
  fmt.Fprintln(out, "}\n")
}
