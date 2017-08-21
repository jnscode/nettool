/*
   IP Protocol header
    0                   1                   2                   3
    0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |Version|  IHL  |Type of Service|          Total Length         |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |         Identification        |Flags|      Fragment Offset    |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |  Time to Live |    Protocol   |         Header Checksum       |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |                       Source Address                          |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |                    Destination Address                        |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |                    Options                    |    Padding    |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

   ICMP Protocol header
    0                   1                   2                   3
    0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |     Type      |     Code      |          Checksum             |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |           Identifier          |        Sequence Number        |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |     Data ...
   +-+-+-+-+-

   Sample:
	param := ping.Param {"www.baidu.com", 32, true, 5}
	r, e := ping.Ping(param)
	if e != nil {
		println("error", e.Error())
	} else {
		fmt.Printf("%v\n", r)
	}
*/

package ping

import (
	"errors"
	"net"
	"os"
	"time"
)

const (
	icmpReqCode = 8
	icmpRepCode = 0

	ipHeadSize   = 24
	icmpHeadSize = 8

	ipTtlPos = 8
)

// struct of ping param
type Param struct {
	Addr    string
	DataLen int
	Segment bool
	Timeout int
}

// struct of ping result
type Result struct {
	Succ  bool
	Ttl   int
	Time  int
	Error error
}

func (r *Result) setError(succ bool, err error) {
	r.Succ = succ
	r.Error = err
}

// calc icmp packet checksum
func checkSum(p []byte) uint16 {
	cklen := len(p)
	s := uint32(0)

	for i := 0; i < (cklen - 1); i += 2 {
		s += uint32(p[i+1])<<8 | uint32(p[i])
	}

	if cklen&1 == 1 {
		s += uint32(p[cklen-1])
	}

	s = (s >> 16) + (s & 0xffff)
	s = s + (s >> 16)

	return uint16(s)
}

// make ping request packet
func makeRequest(id, seq int, dataLen int) []byte {

	dataLen &= 0xff
	size := icmpHeadSize + dataLen
	req := make([]byte, size)

	req[0] = icmpReqCode       // type
	req[1] = 0                 // code
	req[2] = 0                 // cksum
	req[3] = 0                 // cksum
	req[4] = uint8(id >> 8)    // id
	req[5] = uint8(id & 0xff)  // id
	req[6] = uint8(seq >> 8)   // sequence
	req[7] = uint8(seq & 0xff) // sequence

	for i:=8; i < size; i++ {
		req[i] = (byte)('0') + ((byte)(i - 8) % 10)
	}

	// place checksum back in header; using ^= avoids the
	// assumption the checksum bytes are zero
	s := checkSum(req)
	req[2] ^= uint8(^s & 0xff)
	req[3] ^= uint8(^s >> 8)

	return req
}

// parse ping result
func parseResult(p []byte, r *Result) (id, seq int) {

	ipHeadLen := p[0] & 0xf * 4
	if ipHeadLen > ipHeadSize {
		r.setError(false, errors.New("Invalid ip head length"))
		return
	}

	r.Ttl = int(p[ipTtlPos])

	icmpReplyType := p[ipHeadLen]
	if icmpReplyType != icmpRepCode {
		r.setError(false, errors.New("Invalid icmp reply"))
		return
	}

	id = int(p[ipHeadLen+4])<<8 | int(p[ipHeadLen+5])
	seq = int(p[ipHeadLen+6])<<8 | int(p[ipHeadLen+7])

	return
}

// ping addr count times
func Ping(param Param) (Result, error) {

	var result Result

	// *IPAddr
	raddr, e := net.ResolveIPAddr("ip4", param.Addr)
	if e != nil {
		return result, e
	}

	// *IPConn
	conn, e := net.DialIP("ip4:icmp", nil, raddr)
	if e != nil {
		return result, e
	}

	defer conn.Close()

	icmpId := os.Getpid() & 0xffff
	icmpSeq := 1

	req := makeRequest(icmpId, icmpSeq, param.DataLen)

	// send icmp request
	n, e := conn.Write(req)
	if e != nil || n != len(req) {
		result.setError(false, e)
		return result, nil
	}

	// set timeout
	conn.SetDeadline(time.Now().Add(5 * time.Second))

	// make response data buffer
	rsp := make([]byte, ipHeadSize+icmpHeadSize+param.DataLen)

	tmStart := time.Now()

	// read icmp response
	_, e = conn.Read(rsp)
	if e != nil {
		result.setError(false, e)
		return result, nil
	}

	tmCost := time.Since(tmStart)

	// pase result
	rcvid, rcvseq := parseResult(rsp, &result)
	if rcvid != icmpId || rcvseq != icmpSeq {
		result.setError(false, errors.New("icmp id or seq not match"))
	} else {
		result.Time = (int)(tmCost / 1000 / 1000)
		result.Succ = true
	}

	return result, nil
}
