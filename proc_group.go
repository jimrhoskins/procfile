package procfile

import (
  "os/exec"
  "fmt"
  "time"
  "os"
  "strings"
)

const (
  Xcyan = "\033[0;36m"
  Xred = "\033[0;31m"
  Xdefault = "\033[0;00m"
)

type ProcGroup struct {
  Name string
  CmdString string
  Number int
  BasePort int

  Dir string
  PortManager *PortManager

  cmds map[int]*exec.Cmd

  workerCount int
}

func NewProcGroup(name, cmdString string, portManager *PortManager) *ProcGroup{
  return &ProcGroup{
    Name: name,
    CmdString: cmdString,
    PortManager: portManager,
    cmds: make(map[int]*exec.Cmd),
  }
}

func (pg *ProcGroup) Launch() *exec.Cmd {
  return pg.LaunchEnv([]string{})
}


func (pg *ProcGroup) LaunchEnv(env []string) *exec.Cmd {
  processNum := len(pg.cmds)
  processName := fmt.Sprintf("%s.%d", pg.Name, processNum)

  port := pg.BasePort + processNum
  if pg.PortManager != nil {
    port = pg.PortManager.Lease()
  }

  env = append(env, fmt.Sprintf("PORT=%d", port))
  env = append(os.Environ(),  env...)
  envMap := EnvMap(env)

  cmdString := os.Expand(pg.CmdString, func(name string) string {
    return envMap[name]
  })

  args := strings.Fields(cmdString)
  cmd := exec.Command(args[0], args[1:]...)
  cmd.Dir = pg.Dir
  cmd.Env = env




  cmd.Stdout = Prefixer(os.Stdout, func () string {
    return fmt.Sprintf("%s%s %s\t| %s", Xcyan, time.Now().Format("15:04:05"), processName, Xdefault)
  })

  cmd.Stderr = Prefixer(os.Stderr, func () string {
    return fmt.Sprintf("%s%s %s\t| %s", Xred, time.Now().Format("15:04:05"), processName, Xdefault)
  })

  pg.cmds[processNum] = cmd

  cmd.Start()

  return cmd
}

func (pg *ProcGroup) Kill() {
  for i, cmd := range pg.cmds {
    cmd.Process.Kill()
    delete(pg.cmds, i)
  }
}

func (pg *ProcGroup) Addrs() []string {
  ports := make([]string, 0, len(pg.cmds))
  for _, cmd := range pg.cmds {
    port := ""
    for _, expr := range cmd.Env {
      x := strings.SplitN(expr, "=", 2)
      if x[0] == "PORT" {
        port = x[1]
      }
    }

    if port != "" {
      ports = append(ports, fmt.Sprintf(":%s", port))
    }
  }

  return ports

}

func (pg *ProcGroup) Wait() {
  for _, cmd := range pg.cmds {
    cmd.Wait()
  }
}
