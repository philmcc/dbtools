

package network

import (
	"net"
	"strings"
)

func Get_CNAME_details(cname string) (ip_address string, hostname string) {

	//fmt.Println("Checking CNAME: ",cname)
	//var hostname string
	addr, err := net.LookupIP(cname)

	if err != nil {
		//fmt.Println("Unknown host - ",cname, " - Not found!")
		ip_address = "NotFound"
		hostname = "NotFound"
	} else {
		//fmt.Println("IP address: ", addr[0])
		ip_address = addr[0].String()
		host, err := net.LookupAddr(addr[0].String())
		if err != nil {
		//	fmt.Println("Unknown host")
		hostname = "NotFound"
		} else {
		//	fmt.Println("HostName is: ",host[0])
			hostname = host[0]
		}
	}
	hostname = strings.TrimRight(hostname, ".")

	//
	// TODO
	// Store details in node table
	// If node exists update else insert

	return ip_address, hostname
}





/*
func main () {
	cname := "tii-master.sprint.iparadigms.com"
	ip_addr,hostname  := get_details(cname)
	fmt.Println(ip_addr, " ", hostname)

}
*/