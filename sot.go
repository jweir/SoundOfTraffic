package main

// Sound of Traffic (2) server

import (
	"encoding/json"
	"flag"
	"fmt"
	es "github.com/antage/eventsource/http"
	pcap "github.com/miekg/pcap"
	"log"
	"net/http"
)

type addrHdr interface {
	SrcAddr() string
	DestAddr() string
	Len() int
}

type Source struct {
	es    es.EventSource
	label string
	url   string
}

type SourceList map[string]Source

func main() {
	device := flag.String("d", "", "network device")
	flag.Parse()

	// Maps the type (TCP, UDP, ARP, etc) to an event source
	sources := make(SourceList)

	defer sources.add("tcp").Close()
	defer sources.add("udp").Close()

	// TOOD "/protocol" which describes a function to
	// break the stream into sonic data

	http.Handle("/sources", sources)
	http.Handle("/", http.FileServer(http.Dir("./pub")))

	if *device == "" {
		printDevices()
	} else {
		go openDevice(device, sources)
		log.Println("Opening server at localhost:8080 and listening for ", *device)
		log.Fatal(http.ListenAndServe(":8080", nil))
	}

}

// binds an event source to a packet type (ie TCP)
// creates an http handler to allow subscription to the packet type
func (sources SourceList) add(label string) es.EventSource {
	e := es.New(nil)
	c := Source{
		e,
		label,
		urlFor(label),
	}
	http.Handle(urlFor(label), e)
	sources[label] = c
	return e
}

func urlFor(s string) string {
	return "/pcap/" + s + "/"
}

// display the available sources as json
func (c SourceList) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	o := make(map[string]string)
	for k, _ := range c {
		o[k] = c[k].url
	}
	m, _ := json.Marshal(o)
	fmt.Fprintf(w, "%s", m)
}

func openDevice(device *string, sources SourceList) {
	h, err := pcap.OpenLive(*device, 65535, true, 100)
	if h == nil {
		fmt.Printf("Failed to open %s : %s\n", *device, err)
	} else {

		// TODO set the port via an arg
		h.SetFilter("port != 8080")
		for pkt := h.Next(); ; pkt = h.Next() {
			if pkt != nil {
				pkt.Decode()
				sources.send(pkt)
			}
		}

		log.Println("timeout")
	}
}

func (s SourceList) send(pkt *pcap.Packet) {
	if len(pkt.Headers) >= 2 {

		if addr, ok := pkt.Headers[0].(addrHdr); ok {
			switch k := pkt.Headers[1].(type) {
			case *pcap.Tcphdr:
				log.Printf("tcp")
				m := fmt.Sprintf("%s:%d %s:%d", addr.SrcAddr(), int(k.SrcPort), addr.DestAddr(), int(k.DestPort))
				s["tcp"].es.SendMessage(m, "", "")
			case *pcap.Udphdr:
				log.Printf("udp")
				m := fmt.Sprintf("%s:%d %s:%d", addr.SrcAddr(), int(k.SrcPort), addr.DestAddr(), int(k.DestPort))
				s["udp"].es.SendMessage(m, "", "")
			default:
				log.Printf("%T", k)
			}
		}
	}
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
