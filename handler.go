package procfile

import (
  "time"
  "fmt"
  "net/http"
  "github.com/jimrhoskins/soxy/proxy"
  "github.com/jimrhoskins/soxy/balancer"
)

var portmanager = NewPortManager(40000, 50000)

type request struct {
  W http.ResponseWriter
  R *http.Request
  done chan bool
}

func req(w http.ResponseWriter, r *http.Request) *request {
  return &request{w, r, make(chan bool)}
}

type Handler struct {
  Target string
  Dir string

  p *Procfile
  requests chan *request

}

func NewHandler(dir string) *Handler{
  handler := &Handler{
    Dir: dir,
    requests: make(chan *request),
  }

  go func(){
    for {
      req := <-handler.requests
      handler.p = NewProcfile(handler.Dir, portmanager)
      handler.p.SetConcurrency("web=2")
      handler.p.LaunchAll()

      addrs := handler.p.AddrsFor("web")
      if len(addrs) == 0 {
        break
      }

      handlers := make([]http.Handler, 0, len(addrs))
      for _, addr := range addrs {
        <-Listening(addr)
        handlers = append(handlers, proxy.New(addr))
      }

      loadBalancer := balancer.New(handlers...)

      RequestLoop:
      for {
        go func (req *request){
          loadBalancer.ServeHTTP(req.W, req.R)
          req.done <- true
        }(req)

        select {
        case req = <-handler.requests:
          // Continue Loop
        case <-time.After(10 * time.Second):
          fmt.Println("Exiting!!!")
          handler.p.Kill()
          break RequestLoop
        }
      }
    }
  }()

  return handler
}


func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  request := req(w, r)
  h.requests <- request
  <-request.done

}

