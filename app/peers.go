package main

import (
	"fmt"
	"net"
)

type Peer struct {
	IP   net.IP
	Port uint16
}

func (p *Peer) String() string {
	return fmt.Sprintf("%v:%v", p.IP, p.Port)
}
