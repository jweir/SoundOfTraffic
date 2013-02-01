package main

// Sound of Traffic (2) server

import (
	"flag"
	"fmt"
	eventsource "github.com/antage/eventsource/http"
	pcap "github.com/miekg/pcap"
	"log"
	"net/http"
)

func main() {
	device := flag.String("d", "", "network device")
	flag.Parse()

	es := eventsource.New(nil)
	defer es.Close()
	http.Handle("/pcap/", es)
	http.Handle("/", http.FileServer(http.Dir("./pub")))

	if *device == "" {
		printDevices()
	} else {
		go openDevice(device, es)
		log.Println("Opening server at localhost:8080 and listening for ", *device)
		log.Fatal(http.ListenAndServe(":8080", nil))
	}

}

type addrHdr interface {
	SrcAddr() string
	DestAddr() string
	Len() int
}

func openDevice(device *string, es eventsource.EventSource) {
	h, err := pcap.OpenLive(*device, 65535, true, 100)
	if h == nil {
		fmt.Printf("Failed to open %s : %s\n", *device, err)
	} else {

		// TODO this needs to be dynmically set
		h.SetFilter("port != 8080")
		for pkt := h.Next(); ; pkt = h.Next() {
			if pkt != nil {
				pkt.Decode()
        m, ok := prepare(pkt)
        if ok {
          es.SendMessage(m, "", "")
        }
			}
		}

		log.Println("timeout")
	}
}

func prepare(pkt *pcap.Packet) (string, bool) {
	if len(pkt.Headers) >= 2 {
		if t, ok := pkt.Headers[1].(*pcap.Tcphdr); ok {
			if addr, ok := pkt.Headers[0].(addrHdr); ok {
        log.Printf("ok")
				return fmt.Sprintf("%s:%d %s:%d", addr.SrcAddr(), int(t.SrcPort), addr.DestAddr(), int(t.DestPort)), true
			}
		}
	}
	return "", false
}

func printDevices() {
	devs, err := pcap.FindAllDevs()

	if len(devs) == 0 {
		fmt.Printf("Error: no devices found. %s\n", err)
	} else {
		fmt.Println("Available network devices")
		for _, dev := range devs {
			fmt.Printf("  %s \n", dev.Name)
		}
	}
}
