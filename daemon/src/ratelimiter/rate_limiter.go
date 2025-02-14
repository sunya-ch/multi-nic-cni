package ratelimiter

import (
	"fmt"
	"log"
	"math"
	"net"

	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

const (
	HTB_METHOD = "htb"
	TBF_METHOD = "tbf"
)

var (
	rootClass = netlink.MakeHandle(1, 0)
)

type RateLimitRequest struct {
	hostDevice string
	srcIPs     []string
	ratePerIP  float64
}

type RateLimitResponse struct {
	Succeed bool
	Message string
}

// LimitRate limits egress rate for each specific IP range
func LimitRate(req RateLimitRequest) error {

	// Get the interface by name
	link, err := netlink.LinkByName(req.hostDevice)
	if err != nil {
		return fmt.Errorf("Failed to get link: %w", err)
	}
	err = addQdisc(link)
	if err != nil {
		return fmt.Errorf("Failed to add qdisc: %w", err)
	}
	rate := uint64(math.Ceil(req.ratePerIP / 8.0))
	for minor, srcIP := range req.srcIPs {
		if err = addClass(link, rate, uint16(minor)); err != nil {
			return fmt.Errorf("Failed to add class: %w", err)
		}
		if err = addFilter(link, uint16(minor), srcIP); err != nil {
			return fmt.Errorf("Failed to add filter: %w", err)
		}
		log.Printf("Successfully set limit for %s: %d", srcIP, rate)
	}

	return nil
}

func RemoveRateLimit(req RateLimitRequest) error {
	// Get the interface by name
	link, err := netlink.LinkByName(req.hostDevice)
	if err != nil {
		return err
	}
	qdiscList, err := netlink.QdiscList(link)
	if err != nil {
		return err
	}
	for _, qdisc := range qdiscList {
		if qdisc.Type() == "htb" {
			return netlink.QdiscDel(qdisc)
		}
	}
	return nil
}

func addQdisc(link netlink.Link) error {
	// Create a new HTB qdisc (this example uses a simple root qdisc)
	attrs := netlink.QdiscAttrs{
		LinkIndex: link.Attrs().Index,
		Handle:    rootClass,
		Parent:    netlink.HANDLE_ROOT,
	}
	qdisc := netlink.NewHtb(attrs)
	err := netlink.QdiscAdd(qdisc)
	return err
}

func addClass(link netlink.Link, rate uint64, minor uint16) error {
	classId := netlink.MakeHandle(1, minor)
	classattrs := netlink.ClassAttrs{
		LinkIndex: link.Attrs().Index,
		Handle:    classId,
		Parent:    rootClass,
	}

	bytespersecond := math.Ceil(float64(rate) / 8.0)
	var class netlink.Class
	htbclassattrs := netlink.HtbClassAttrs{
		Rate:    rate,
		Buffer:  uint32(bytespersecond/netlink.Hz() + float64(link.Attrs().MTU) + 1),
		Cbuffer: uint32(bytespersecond/netlink.Hz() + 10*float64(link.Attrs().MTU) + 1),
	}
	class = netlink.NewHtbClass(classattrs, htbclassattrs)
	return netlink.ClassAdd(class)
}

func addFilter(link netlink.Link, minor uint16, srcIP string) error {
	classId := netlink.MakeHandle(1, minor)

	// Convert the source IP string to net.IP type
	ip := net.ParseIP(srcIP)
	if ip == nil {
		return fmt.Errorf("failed to parse IP: %s", ip)
	}

	// Create the u32 selector
	// Match source IP in the IP header (u32 selector)
	// Offset 12 corresponds to the IP header's source address
	mask := uint32(0xFFFFFFFF) // Full 32-bit match
	value := uint32(ip[12])<<24 | uint32(ip[13])<<16 | uint32(ip[14])<<8 | uint32(ip[15])

	// Create the u32 selector (filter)
	u32Sel := &netlink.TcU32Sel{
		Flags:    0,      // Set appropriate flags
		Offshift: 0,      // No shift needed
		Nkeys:    1,      // We're only matching 1 key (source IP)
		Pad:      0,      // Padding for alignment
		Offmask:  0xFFFF, // Mask for the whole 32-bit source IP field
		Off:      12,     // Offset to the source IP address in the IPv4 header
		Offoff:   0,      // No offset for the key
		Hoff:     0,      // No header offset needed
		Hmask:    0,      // No header mask
		Keys: []netlink.TcU32Key{
			{
				Val:  value, // IP in hex
				Mask: mask,  // Match the whole IP address
			},
		},
	}

	filter := &netlink.U32{
		FilterAttrs: netlink.FilterAttrs{
			LinkIndex: link.Attrs().Index,
			Parent:    rootClass,
			Priority:  1,
			Protocol:  unix.ETH_P_IP,
		},
		ClassId: classId,
		Sel:     u32Sel,
	}
	return netlink.FilterAdd(filter)
}
