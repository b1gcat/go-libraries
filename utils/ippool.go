package utils

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"sync"
)

type IpPool struct {
	lock sync.RWMutex

	Next  net.IP
	Start net.IP
	End   net.IP
}

func (p *IpPool) Alloc() (ip net.IP, err error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if bytes.Compare(p.Next, p.End) == 0 {
		return nil, fmt.Errorf("无可分配的IP")
	}
	ip = p.Next
	ipInt := binary.BigEndian.Uint32(p.Next.To4()) + 1
	binary.BigEndian.PutUint32(p.Next, ipInt)
	return
}
