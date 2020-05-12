package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/miekg/dns"
)

//
// Query - DNS query structure
//
type Query struct {
	qname  string
	qtype  uint16
	qclass uint16
}

//
// Set query components
//
func (q *Query) Set(qname string, qtype uint16, qclass uint16) {
	q.qname = qname
	q.qtype = qtype
	q.qclass = qclass
}

//
// QueryOptions - query options
//
type QueryOptions struct {
	rdflag  bool
	adflag  bool
	cdflag  bool
	timeout time.Duration
	retries int
	payload uint16
}

//
// getQuery - get populated Query struct
//
func getQuery(qname string, qtype uint16, qclass uint16) Query {
	var query Query
	query.Set(qname, qtype, qclass)
	return query
}

//
// AddressString - compose address string for net functions
//
func addressString(ipaddress net.IP, port int) string {
	addr := ipaddress.String()
	if strings.Index(addr, ":") == -1 {
		return addr + ":" + strconv.Itoa(port)
	}
	return "[" + addr + "]" + ":" + strconv.Itoa(port)
}

//
// GetResolver - obtain (1st) system default resolver address
//
func getResolver() (resolver net.IP, err error) {
	config, err := dns.ClientConfigFromFile("/etc/resolv.conf")
	if err == nil {
		resolver = net.ParseIP(config.Servers[0])
	}
	return resolver, err
}

//
// MakeQuery - construct a DNS query MakeMessage
//
func makeQuery(query Query, qopts QueryOptions) *dns.Msg {
	m := new(dns.Msg)
	m.Id = dns.Id()
	if qopts.rdflag {
		m.RecursionDesired = true
	} else {
		m.RecursionDesired = false
	}
	if qopts.adflag {
		m.AuthenticatedData = true
	} else {
		m.AuthenticatedData = false
	}
	if qopts.cdflag {
		m.CheckingDisabled = true
	} else {
		m.CheckingDisabled = false
	}
	m.SetEdns0(qopts.payload, true)
	m.Question = make([]dns.Question, 1)
	m.Question[0] = dns.Question{Name: query.qname, Qtype: query.qtype,
		Qclass: query.qclass}
	return m
}

//
// SendQueryUDP - send DNS query via UDP
//
func sendQueryUDP(query Query, resolver net.IP, qopts QueryOptions) (*dns.Msg, error) {

	var response *dns.Msg
	var err error

	destination := addressString(resolver, 53)

	m := makeQuery(query, qopts)

	c := new(dns.Client)
	c.Net = "udp"
	c.Timeout = qopts.timeout

	retries := qopts.retries
	for retries > 0 {
		response, _, err = c.Exchange(m, destination)
		if err == nil {
			break
		}
		if nerr, ok := err.(net.Error); ok && !nerr.Timeout() {
			break
		}
		retries--
	}

	return response, err
}

//
// SendQueryTCP - send DNS query via TCP
//
func sendQueryTCP(query Query, resolver net.IP, qopts QueryOptions) (*dns.Msg, error) {

	var response *dns.Msg
	var err error

	destination := addressString(resolver, 53)
	m := makeQuery(query, qopts)

	c := new(dns.Client)
	c.Net = "tcp"
	c.Timeout = qopts.timeout

	response, _, err = c.Exchange(m, destination)
	return response, err

}

//
// SendQuery - send DNS query via UDP with fallback to TCP upon truncation
//
func sendQuery(query Query, resolver net.IP, qopts QueryOptions) (*dns.Msg, error) {

	var response *dns.Msg
	var err error

	response, err = sendQueryUDP(query, resolver, qopts)

	if err == nil && response.MsgHdr.Truncated {
		response, err = sendQueryTCP(query, resolver, qopts)
	}

	if err != nil {
		return nil, err
	}
	if response == nil {
		return nil, fmt.Errorf("Error: null DNS response to query")
	}
	switch response.MsgHdr.Rcode {
	case dns.RcodeSuccess:
		break
	case dns.RcodeNameError:
		return nil, fmt.Errorf("NXDOMAIN: %s doesn't exist", query.qname)
	default:
		return nil, fmt.Errorf("Error: Response code: %s",
			dns.RcodeToString[response.MsgHdr.Rcode])
	}

	var rrcount int
	for _, rr := range response.Answer {
		if rr.Header().Rrtype == query.qtype {
			rrcount++
		}
	}
	if rrcount == 0 {
		return nil, fmt.Errorf("NODATA: %s/%d", query.qname, query.qtype)
	}

	return response, err

}
