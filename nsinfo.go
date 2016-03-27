package main

import (
	"os"
	"fmt"
	"strings"
	"net"
	"sort"
	"github.com/miekg/dns"
)


func usage() {
	fmt.Printf("Usage: nsinfo <zonename>\n")
}


func mk_fqdn(qname string) (string) {
	if strings.HasSuffix(qname, ".") {
		return qname
	} else {
		return qname + "."
	}
}


func do_query(qname, qtype string) (response *dns.Msg) {

	m := new(dns.Msg)
	m.Id = dns.Id()
	m.RecursionDesired = true
	// m.SetEdns0(4096, true)
	m.Question = make([]dns.Question, 1)
	qtype_int, ok := dns.StringToType[qtype]
	if !ok {
		fmt.Printf("%s: Unrecognized query type.\n", qtype)
		return nil
	}
	m.Question[0] = dns.Question{qname, qtype_int, dns.ClassINET}

	c := new(dns.Client)
	response, _, err := c.Exchange(m, "127.0.0.1:53")
	if err != nil {
		fmt.Printf("DNS query failed %s\n", err);
		return nil
	}
	switch response.MsgHdr.Rcode {
	case dns.RcodeSuccess:
		break
	case dns.RcodeNameError:
		fmt.Printf("NXDOMAIN: %s: name doesn't exist\n", qname)
		return nil
	default:
		fmt.Printf("Error: Response code: %s\n", 
			dns.RcodeToString[response.MsgHdr.Rcode])
		return nil
	}

	var rrcount int
	for _, rr := range response.Answer {
		if rr.Header().Rrtype == qtype_int {
			rrcount += 1
		}
        }
	if rrcount == 0 {
		/* fmt.Printf("NODATA: %s/%s\n", qname, qtype) */
		return nil
	}

	return response
}


func getIPAddresses(hostname, rrtype string) ([]net.IP) {

	var a_rr *dns.A
        var aaaa_rr *dns.AAAA
	var ip_list []net.IP

	switch rrtype {

	case "AAAA":
		response := do_query(hostname, rrtype)
		if response != nil {
			for _, rr := range response.Answer {
				if rr.Header().Rrtype == dns.TypeAAAA {
					aaaa_rr = rr.(*dns.AAAA)
					ip_list = append(ip_list, aaaa_rr.AAAA)
				}
			}
		
		}
	case "A":
		response := do_query(hostname, rrtype)
		if response != nil {
			for _, rr := range response.Answer {
				if rr.Header().Rrtype == dns.TypeA {
					a_rr = rr.(*dns.A)
					ip_list = append(ip_list, a_rr.A)
				}
			}
		
		}
	default:
		fmt.Printf("getIPAddresses: %s: invalid rrtype\n", rrtype)
	}

	return ip_list

}


func reverseLookup(ipaddr net.IP) (string) {

	arpaname, err := dns.ReverseAddr(ipaddr.String())
	if err != nil {
		return ""
	}
	response := do_query(arpaname, "PTR")
	if response == nil {
		return ""
	}
	if len(response.Answer) < 1 {
		return ""
	}
	ptr_rr := response.Answer[0].(*dns.PTR)
	return ptr_rr.Ptr
}


const hexDigit = "0123456789abcdef"
var ip2asnSuffixV4 = "origin.asn.cymru.com."
var ip2asnSuffixV6 = "origin6.asn.cymru.com."

func ip2asn(ip net.IP) (string) {

	var h1, h2 byte
	var qname string

	if ip.To4() != nil {
		// IPv4 address
		qname = fmt.Sprintf("%d.%d.%d.%d.%s",
			int(ip[15]),
			int(ip[14]),
			int(ip[13]),
			int(ip[12]),
			ip2asnSuffixV4)
	} else {
		// IPv6 address
		buf6 := make([]byte, 0, len(ip)*4)
		for i := len(ip)-1; i >= 0; i-- {
			h1 = ip[i] & 0xf
			h2 = ip[i] >> 4
			buf6 = append(buf6, hexDigit[h1])
			buf6 = append(buf6, '.')
			buf6 = append(buf6, hexDigit[h2])
			buf6 = append(buf6, '.')
		}
		qname = string(buf6) + ip2asnSuffixV6
	}

	response := do_query(qname, "TXT")
	if response == nil {
		fmt.Printf("No ASN found\n")
		return ""
	}

        txt_rr := response.Answer[0].(*dns.TXT)
	return "AS" + strings.Split(txt_rr.Txt[0], " ")[0]

}


func main() {

	if len(os.Args) != 2 {
		usage()
		return
	}
	zone := mk_fqdn(os.Args[1])
	response := do_query(zone, "NS")
	if response == nil {
		return
	}

	var ns_rr *dns.NS
	var ns_name_list []string

	for _, rr := range response.Answer {
		if rr.Header().Rrtype == dns.TypeNS {
			ns_rr = rr.(*dns.NS)
			ns_name_list = append(ns_name_list, ns_rr.Ns)
		}
	}

	// should really be sorting in DNS canonical order here .. //
	sort.Strings(ns_name_list)

	var v4_list []net.IP
	var v6_list []net.IP

	for _, ns_name := range ns_name_list {
		v4_list = nil
		v6_list = nil

		v6_list = getIPAddresses(ns_name, "AAAA")
		for _, ip := range v6_list {
			fmt.Printf("%s %s %s %s\n", ns_name, ip, ip2asn(ip), reverseLookup(ip))
		}


		v4_list = getIPAddresses(ns_name, "A")
		for _, ip := range v4_list {
			fmt.Printf("%s %s %s %s\n", ns_name, ip, ip2asn(ip), reverseLookup(ip))
		}

	}

}
