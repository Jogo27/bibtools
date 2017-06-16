package main

import "bufio"
import "fmt"
import "os"
import "regexp"

func ReadAuxFile(filename string) map[string] bool {

  sent := make(map[string] bool)

  var file *os.File
  var err  error
  if file, err = os.Open(filename) ; err != nil {
    fmt.Fprintf(os.Stderr, "Unable to read %s\n", filename);
    return sent
  }

  scanner := bufio.NewScanner(file)
  rex := regexp.MustCompile("\\\\citation{([^\n}]+)}");
  for scanner.Scan() {
    if match := rex.FindSubmatch(scanner.Bytes()) ; match != nil {
      sent[string(match[1])] = true
    }
  }

  file.Close()
  return sent
}

func main() {
  if len(os.Args) < 3 {
    fmt.Fprintf(os.Stderr, "Too few arguments\n");
    return;
  }


  chbib := make(chan BibElement, 32)
  go ReadBibFile(os.Args[2], chbib)

  refset := ReadAuxFile(os.Args[1])

  for {
    elem, ok := <-chbib
    if !ok { break }
    if refset[elem.name] {
      elem.Display(os.Stdout)
    }
  }
}
