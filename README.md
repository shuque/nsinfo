# nsinfo
Print info about a zone's name servers

nsinfo

For a given DNS zone, print out some information about its authoritative
nameservers, specifically nameserver name, IP address, ASN originating
prefix covering the IP address, and reverse mapping of IP address to name.

Author: Shumon Huque <shuque@gmail.com>

Example runs:

```
$ nsinfo complex.com
ns1.complex.com. 2606:2800:3::5 AS[15133] ns1.edgecastdns.net.
ns1.complex.com. 192.16.16.5 AS[15133] ns1.edgecastdns.net.
ns2.complex.com. 2606:2800:3::6 AS[15133] ns2.edgecastdns.net.
ns2.complex.com. 192.16.16.6 AS[15133] ns2.edgecastdns.net.
ns3.complex.com. 2606:2800:c::5 AS[15133] ns3.edgecastdns.net.
ns3.complex.com. 198.7.29.5 AS[15133] ns3.edgecastdns.net.
ns4.complex.com. 2606:2800:c::6 AS[15133] ns4.edgecastdns.net.
ns4.complex.com. 198.7.29.6 AS[15133] ns4.edgecastdns.net.

$ nsinfo upenn.edu
adns1.upenn.edu. 2607:f470:1001::1:a AS[55] adns1.upenn.edu.
adns1.upenn.edu. 128.91.3.128 AS[55] adns1.upenn.edu.
adns2.upenn.edu. 2607:f470:1002::2:3 AS[55] adns2.upenn.edu.
adns2.upenn.edu. 128.91.254.22 AS[55] adns2.upenn.edu.
adns3.upenn.edu. 2607:f470:1003::3:c AS[55] adns3.upenn.edu.
adns3.upenn.edu. 128.91.251.33 AS[55] adns3.upenn.edu.
dns1.udel.edu. 128.175.13.16 AS[34] dns1.udel.edu.
dns2.udel.edu. 128.175.13.17 AS[34] dns2.udel.edu.
sns-pb.isc.org. 2001:500:2e::1 AS[3557] sns-pb.isc.org.
sns-pb.isc.org. 192.5.4.1 AS[3557] sns-pb.isc.org.

$ nsinfo com
a.gtld-servers.net. 2001:503:a83e::2:30 AS[36622] a.gtld-servers.net.
a.gtld-servers.net. 192.5.6.30 AS[36617 36619 36620 36625] a.gtld-servers.net.
b.gtld-servers.net. 2001:503:231d::2:30 AS[26415] b.gtld-servers.net.
b.gtld-servers.net. 192.33.14.30 AS[26415] b.gtld-servers.net.
c.gtld-servers.net. 192.26.92.30 AS[36617 36619 36620 36625] c.gtld-servers.net.
d.gtld-servers.net. 192.31.80.30 AS[36617 36619 36620 36625] d.gtld-servers.net.
e.gtld-servers.net. 192.12.94.30 AS[36618 36622 36624 36628] e.gtld-servers.net.
f.gtld-servers.net. 192.35.51.30 AS[36618 36622 36624 36628] f.gtld-servers.net.
g.gtld-servers.net. 192.42.93.30 AS[36618 36622 36624 36628] g.gtld-servers.net.
h.gtld-servers.net. 192.54.112.30 AS[36621 36623 36632] h.gtld-servers.net.
i.gtld-servers.net. 192.43.172.30 AS[36621 36623 36632] i.gtld-servers.net.
j.gtld-servers.net. 192.48.79.30 AS[36621 36623 36632] j.gtld-servers.net.
k.gtld-servers.net. 192.52.178.30 AS[36631] k.gtld-servers.net.
l.gtld-servers.net. 192.41.162.30 AS[36619] l.gtld-servers.net.
m.gtld-servers.net. 192.55.83.30 AS[36627] m.gtld-servers.net.
```
