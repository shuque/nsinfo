# nsinfo
Print info about a zone's name servers

nsinfo.py

For a given DNS zone, print out some information about its authoritative
nameservers, specifically nameserver name, IP address, ASN originating
prefix covering the IP address, and reverse mapping of IP address to name.

Author: Shumon Huque <shuque@gmail.com>

Example runs:

```
$ nsinfo.py complex.com
ns1.complex.com. 192.16.16.5 AS15133 ns1.edgecastdns.net.
ns1.complex.com. 2606:2800:3::5 AS15133 ns1.edgecastdns.net.
ns2.complex.com. 192.16.16.6 AS15133 ns2.edgecastdns.net.
ns2.complex.com. 2606:2800:3::6 AS15133 ns2.edgecastdns.net.
ns3.complex.com. 198.7.29.5 AS15133 ns3.edgecastdns.net.
ns3.complex.com. 2606:2800:c::5 AS15133 ns3.edgecastdns.net.
ns4.complex.com. 198.7.29.6 AS15133 ns4.edgecastdns.net.
ns4.complex.com. 2606:2800:c::6 AS15133 ns4.edgecastdns.net.

$ nsinfo.py upenn.edu
dns1.udel.edu. 128.175.13.16 AS34 dns1.udel.edu.
dns2.udel.edu. 128.175.13.17 AS34 dns2.udel.edu.
adns1.upenn.edu. 128.91.3.128 AS55 adns1.upenn.edu.
adns1.upenn.edu. 2607:f470:1001::1:a AS55 adns1.upenn.edu.
adns2.upenn.edu. 128.91.254.22 AS55 adns2.upenn.edu.
adns2.upenn.edu. 2607:f470:1002::2:3 AS55 adns2.upenn.edu.
adns3.upenn.edu. 128.91.251.33 AS55 adns3.upenn.edu.
adns3.upenn.edu. 2607:f470:1003::3:c AS55 adns3.upenn.edu.
sns-pb.isc.org. 192.5.4.1 AS3557 sns-pb.isc.org.
sns-pb.isc.org. 2001:500:2e::1 AS3557 sns-pb.isc.org.
```
