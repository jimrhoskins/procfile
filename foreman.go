package procfile

import (
  "io"
  "os"
  "bufio"
  "fmt"
  "strconv"
  "strings"
)

// Procfile is a collection of process definitions
type Procfile struct {
  Dir           string
  ProcGroups    map[string]*ProcGroup
  PortManager   *PortManager
  File          string
  BasePort      int
  EnvFiles      string

  concurrency   map[string] int

}

type StartArgs struct {
  Directory string
  Procfile string
  EnvFiles string
  Concurrency string
  Port int
}

func Start(args StartArgs){
  p := &Procfile{
    Dir: args.Directory,
    File: args.Procfile,
    BasePort: args.Port,
    EnvFiles: args.EnvFiles,
    ProcGroups: make(map[string] *ProcGroup),
  }

  if len(args.Concurrency) >0 {
    p.SetConcurrency(args.Concurrency)
  }

  p.Parse()
  p.LaunchAll()
  p.Wait()
}


func NewProcfile(dir string, portManager *PortManager) *Procfile {
  p := &Procfile{
    Dir: dir,
    PortManager: portManager,
    ProcGroups: make(map[string] *ProcGroup),
  }

  p.Parse()
  return p
}

func (p *Procfile) SetConcurrency(concurrency string) error {
  p.concurrency = make(map[string] int)


  for _, pair := range strings.Split(concurrency, ",") {
    tmp := strings.SplitN(pair, "=", 2)
    name := strings.TrimSpace(tmp[0])
    num, err := strconv.ParseInt(tmp[1], 10, 0)
    if err != nil {
      return err
    }
    p.concurrency[name] = int(num)
  }

  return nil
}

// Launch one instance of each group
func (p *Procfile) LaunchAll() {
  if len(p.concurrency) == 0 {
    for _, group := range p.ProcGroups {
      group.Launch()
    }
  } else {
    for name, group := range p.ProcGroups {
      n, ok := p.concurrency[name]
      if ok {
        for i := 0; i < n; i++ {
          group.Launch()
        }
      }
    }
  }
}

// Launch one new process from the named group
func (p *Procfile) Launch(name string) {
  group, ok := p.ProcGroups[name]
  if ok {
    group.Launch()
  }
}

// Wait on all processes to complete
func (p *Procfile) Wait() {
  for _, group := range p.ProcGroups {
    group.Wait()
  }
}

// Kill all process in the Procfile
func (p *Procfile) Kill() {
  for _, group := range p.ProcGroups {
    group.Kill()
  }
}

func (p *Procfile) AddrsFor(group string) []string {
  g, ok := p.ProcGroups[group]
  if ok {
    return g.Addrs()
  }
  return []string{}
}

// Parse reads the procfile definition and creates the ProcGroups
// defined in it
func (p *Procfile) Parse() error {
  filename := fmt.Sprintf("%s/Procfile", p.Dir)
  file, err := os.Open(filename)
  if err != nil {
    return err
  }
  defer file.Close()
  reader := bufio.NewReader(file)

  group_number := 0
  for {
    line, err := reader.ReadString('\n')
    if err == io.EOF {
      break
    } else if err != nil {
      return err
    }

    parts := strings.SplitN(line, ":", 2)
    name := strings.TrimSpace(parts[0])
    cmdString := strings.TrimSpace(parts[1])


    group := NewProcGroup(name, cmdString, p.PortManager)
    group.BasePort = p.BasePort + (100 * group_number)
    group.Dir = p.Dir
    group.Number = group_number

    p.ProcGroups[name] = group
    group_number++
  }
  return nil
}

// EnvMap takes a slice of strings in the form of "NAME=value"
// and returns a map of NAME to value. Duplicate names hold
// the last value defined
func EnvMap(env []string) map[string] string{
  envMap := make(map[string] string)

  for _, expr := range env {
    x := strings.SplitN(expr, "=", 2)
    if len(x) == 2 {
      envMap[x[0]] = x[1]
    }
  }

  return envMap
}


// PrefixedWriter write to Destination and prefix each line
// (delimited by '\n') with the result of Prefix
// of Prefix
type PrefixedWriter struct {
  Prefix func () string
  Destination io.Writer

  inLine bool
}

// Prefixer creates a PrefixedWriter
func Prefixer(dst io.Writer, prefix func() string) io.Writer {
  return &PrefixedWriter{
    Prefix: prefix,
    Destination: dst,
  }
}

// Writes to Dest while applying the prefix at each line
func (pw* PrefixedWriter) Write(data []byte) (n int, err error) {
  for _, b := range data {
    if !pw.inLine {
      io.WriteString(pw.Destination, pw.Prefix())
    }
    pw.Destination.Write([]byte{b})

    pw.inLine = b != '\n'
  }

  return len(data), nil
}
