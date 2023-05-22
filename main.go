package main

import (
  "bufio"
  "bytes"
  "flag"
  "fmt"
  "io"
  "os"
  "regexp"
  "strconv"
  "strings"
  "time"
)

var (
  lineMax      = 256 * 1024
  outfile      string
  size         string
  keep         int
  dated        bool
  sizeLimit    int64
  FH           *os.File
  bytesWritten int64
)

func main() {
  flag.StringVar(&outfile, "out", "", "Output file")
  flag.StringVar(&size, "size", "100M", "Size limit")
  flag.IntVar(&keep, "keep", 20, "Files to keep")
  flag.BoolVar(&dated, "dated", false, "Prepend date")
  flag.Parse()

  if outfile == "" {
    fmt.Println("-out <outfile> is required")
    return
  }

  var err error
  sizeLimit, err = calcSize(size)
  if err != nil {
    fmt.Println("-size <bytes> must be positive!")
    return
  }

  if keep <= 0 {
    fmt.Println("-keep <files_to_keep> must be positive")
    return
  }

  openFile(outfile)
  pipe()
}

func openFile(file string) error {
  if FH != nil {
    FH.Close()
  }
  currentSize, _ := getFileSize(file)
  bytesWritten = currentSize
  f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
  if err != nil {
    fmt.Printf("Can't open %s for appending: %s\n", file, err)
    os.Exit(1)
  }
  FH = f
  return nil
}

func rotateFiles() {
  for i := keep - 1; i > 0; i-- {
    currentFile := fmt.Sprintf("%s.%d", outfile, i)
    nextFile := fmt.Sprintf("%s.%d", outfile, i+1)
    if _, err := os.Stat(currentFile); err == nil {
      if err := os.Rename(currentFile, nextFile); err != nil {
        fmt.Printf("Rename failed %s -> %s: %s\n", currentFile, nextFile, err)
        os.Exit(1)
      }
    }
  }
  if err := os.Rename(outfile, fmt.Sprintf("%s.1", outfile)); err != nil {
    fmt.Printf("Can't rename %s to %s.1: %s\n", outfile, outfile, err)
    os.Exit(1)
  }
  openFile(outfile)
}

func append(data []byte) {
  checkRotate(len(data))
  nwritten, err := FH.Write(data)
  if err != nil {
    fmt.Printf("Error writing to file: %s\n", err)
    os.Exit(1)
  }
  bytesWritten += int64(nwritten)
}

func checkRotate(argLength int) {
  if sizeLimit <= 0 {
    return
  }

  // don't rotate on empty files
  if bytesWritten > 0 && bytesWritten+int64(argLength) > sizeLimit {
    rotateFiles()
  }
}

func pipe() {
  buffer := make([]byte, 256*1024)
  for {
    nRead, err := os.Stdin.Read(buffer)
    if err == io.EOF {
      break
    }
    buf := buffer[:nRead]
    if dated {
      scanner := bufio.NewScanner(bytes.NewReader(buf))
      line := ""
      dt := time.Now().Format(time.RFC3339)
      for scanner.Scan() {
        line = scanner.Text()
        line = fmt.Sprintf("%s %s\n", dt, line)
        append([]byte(line))
      }

    } else {
      append(buf)
    }
  }

  FH.Close()
}

func calcSize(arg string) (int64, error) {
  unitLookup := map[string]int64{
    "k":   1000,
    "K":   1024,
    "m":   1000000,
    "M":   1048576,
    "g":   1000000000,
    "G":   1073741824,
    "kB":  1000,
    "KB":  1024,
    "KiB": 1024,
    "mB":  1000000,
    "MB":  1048576,
    "MiB": 1048576,
    "gB":  1000000000,
    "GB":  1073741824,
    "GiB": 1073741824,
  }

  r := regexp.MustCompile(`^([\d_,]+)(\D+)?$`)
  matches := r.FindStringSubmatch(arg)
  if matches == nil {
    return 0, fmt.Errorf("not valid size: %s. Expect a valid size notation e.g. '10K', '100000', '30g', '1,000,000', '4_000_000k'", arg)
  }

  countStr := strings.ReplaceAll(matches[1], "_", "")
  count, err := strconv.ParseInt(countStr, 10, 64)
  if err != nil {
    return 0, err
  }

  unit := matches[2]
  if unit == "" {
    return count, nil
  }

  lookup, ok := unitLookup[unit]
  if !ok {
    return 0, fmt.Errorf("invalid unit %s. Only support %s", unit, strings.Join(keys(unitLookup), ","))
  }

  return count * lookup, nil
}

func keys(m map[string]int64) []string {
  keys := make([]string, len(m))
  i := 0
  for k := range m {
    keys[i] = k
    i++
  }
  return keys
}

func getFileSize(file string) (int64, error) {
  info, err := os.Stat(file)
  if err != nil {
    return 0, err
  }
  return info.Size(), nil
}
