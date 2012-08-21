package main

import (
  "fmt"
  "procfile"
  "os"
  "path"
  "flag"
)

var (
  directory = flag.String("d", "", "root directory containing Procfile")
  env = flag.String("e", ".env", "Specify an alternative envoronment file: -e file1,file2")
  proc = flag.String("f", "", "Specify and alternate procfile location. The containing directory will override -d")
  concur = flag.String("c", "", "Specify the number of each process type to run: -p")
  port = flag.Int("p", 4000, "Specify the base port of the application. Should be a multiple of 1000")

)
func main () {
  flag.Parse()
  var err error

  if *directory == "" {
    *directory, err = os.Getwd()
    if err != nil {
      panic(err)
    }
  }

  if *proc == "" {
    *proc = "Procfile"
  } else {
    *directory = path.Dir(*proc)
    *proc = path.Base(*proc)
  }

  if len(flag.Args()) < 1 {
    fmt.Println("Derp")
    return
  }

  if flag.Arg(0) == "start" {
    procfile.Start(procfile.StartArgs{
      Directory: *directory,
      Procfile: *proc,
      EnvFiles: *env,
      Concurrency: *concur,
      Port: *port,
    })
  }


}
