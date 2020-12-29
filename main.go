package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
)

var errInvalidIP = errors.New("invalid IP address")

func ipToUint32(s string) (uint32, error) {
	ip := net.ParseIP(s)
	if ip == nil {
		return 0, errInvalidIP
	}

	ip = ip.To4()
	if ip == nil {
		return 0, errInvalidIP
	}

	return uint32(ip[3]) | uint32(ip[2])<<8 | uint32(ip[1])<<16 | uint32(ip[0])<<24, nil
}

func uint32ToIP(n uint32) net.IP {
	return net.IPv4(byte(n>>24), byte(n>>16&0xFF), byte(n>>8)&0xFF, byte(n&0xFF))
}

// both side inclusive
func ipRangeToCIDR(start, end uint32) []string {
	if start > end {
		return nil
	}

	// use uint64 to prevent overflow
	ip := int64(start)
	tail := int64(0)
	cidr := make([]string, 0)

	// decrease mask bit
	for {
		// count number of tailing zero bits
		for ; tail < 32; tail++ {
			if (ip>>(tail+1))<<(tail+1) != ip {
				break
			}
		}
		if ip+(1<<tail)-1 > int64(end) {
			break
		}
		cidr = append(cidr, fmt.Sprintf("%s/%d", uint32ToIP(uint32(ip)).String(), 32-tail))
		ip += 1 << tail
	}

	// increase mask bit
	for {
		for ; tail >= 0; tail-- {
			if ip+(1<<tail)-1 <= int64(end) {
				break
			}
		}
		if tail < 0 {
			break
		}
		cidr = append(cidr, fmt.Sprintf("%s/%d", uint32ToIP(uint32(ip)).String(), 32-tail))
		ip += 1 << tail
		if ip-1 == int64(end) {
			break
		}
	}

	return cidr
}

func main() {
	exclusive := flag.Bool("e", false, "exclude end ip")
	flag.Parse()

	if len(flag.Args()) < 2 {
		log.Fatal("Usage: ip-range-to-cidr <start-ip> <end-ip>")
	}

	start, err := ipToUint32(flag.Args()[0])
	if err != nil {
		log.Fatal(err)
	}

	end, err := ipToUint32(flag.Args()[1])
	if err != nil {
		log.Fatal(err)
	}

	if *exclusive {
		if end == 0 {
			os.Exit(0)
		}
		end--
	}

	cidr := ipRangeToCIDR(start, end)
	for _, s := range cidr {
		fmt.Println(s)
	}
}
