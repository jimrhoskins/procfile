package procfile

import (
	"fmt"
	"net"
	"time"
)

func PortAvailable(port int) bool {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Println(err)
		return false
	}
	l.Close()
	return true
}

// Listening returns a channel that will provide a value when addr is
// reachable
func Listening(addr string) <-chan int {
	connected := make(chan int)
	go func() {
		for {
			conn, err := net.Dial("tcp", addr)
			if err != nil {
				time.Sleep(200 * time.Millisecond)
			} else {
				conn.Close()
				connected <- 1
				return
			}
		}
	}()
	return connected
}

type PortManager struct {
	nextPort chan int
	MaxPort  int
	MinPort  int
}

func (p *PortManager) Lease() (port int) {
	for {
		port = <-p.nextPort
		if PortAvailable(port) {
			return port
		}
	}
	return 0
}

func NewPortManager(start, end int) *PortManager {
  next := make(chan int)

	p := &PortManager{
		nextPort: next,
		MinPort:  start,
		MaxPort:  end,
	}

	go func() {
		i := p.MinPort
		for {
			next <- i

			i++
			if i > p.MaxPort {
				i = p.MinPort
			}
		}
	}()

	return p
}
