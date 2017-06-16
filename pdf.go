package main

/* Info/
    Author
    Title
    Subject
    doi
    Keywords
*/

import "fmt"
import "os"
import pdf "rsc.io/pdf"

func printValueRec(val pdf.Value, indent int) {
  switch val.Kind() {
    case pdf.Null:
      fmt.Printf("Null")
    case pdf.Bool:
      fmt.Printf("%t", val.Bool())
    case pdf.Integer:
      fmt.Printf("%d", val.Int64())
    case pdf.Real:
      fmt.Printf("%f", val.Float64())
    case pdf.String:
      fmt.Printf("%s", val.Text())
    case pdf.Name:
      fmt.Printf("%s", val.Name())
    case pdf.Dict:
      fmt.Printf("{\n")
      nindent := indent + 2
      for _, key := range val.Keys() {
        fmt.Printf("%*s%s => ", nindent, "", key)
        printValueRec(val.Key(key), nindent + 4)
        fmt.Println(",")
      }
      fmt.Printf("%*s}", indent, "")
    case pdf.Array:
      fmt.Printf("Array")
    case pdf.Stream:
      fmt.Printf("Stream")
    default:
      fmt.Printf("Unknown")
  }
}

func printValue(val pdf.Value) {
  printValueRec(val, 0)
  fmt.Printf("\n")
}

func displayOutlineTree(root pdf.Outline, indent int) {
  fmt.Printf("%*s%s\n", indent, "", root.Title)
  for _, child := range root.Child {
    displayOutlineTree(child, indent + 2)
  }
}

func main() {
  if len(os.Args) < 2 {
    fmt.Fprintf(os.Stderr, "Too few arguments\n");
    return;
  }

  file, err := pdf.Open(os.Args[1])
  if err != nil {
    fmt.Fprintf(os.Stderr, "%v\n", err)
    return
  }

  printValue(file.Trailer())
  displayOutlineTree(file.Outline(), 0)
}
