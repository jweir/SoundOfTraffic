package main

// Sound of Traffic (2) server

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	es "github.com/antage/eventsource"
	pcap "github.com/miekg/pcap"
)

func main() {
	device := flag.String("i", "", "network device")
	port := flag.String("p", "8000", "http server port (default 8000)")

	flag.Parse()

	if *device == "" {
		flag.PrintDefaults()
		printDevices()
	} else {
		startServer(device, port)
	}
}

// Prints a list of available network devices to console
func printDevices() {
	devs, err := pcap.FindAllDevs()
	if len(devs) == 0 {
		fmt.Printf("Error: no network devices found. %s\n", err)
	} else {
		fmt.Println("Available network devices")
		for _, dev := range devs {
			fmt.Printf("  %s \n", dev.Name)
		}
	}
}

func startServer(device *string, port *string) {
	sm := make(SourceMap)

	for _, t := range []string{"tcp", "udp"} {
		defer sm.Add(t).Close()
		http.Handle(sm[t].url, sm[t].es)
	}

	http.Handle("/sources", sm)
	http.Handle("/", http.FileServer(http.Dir("./pub")))

	go openDevice(device, port, sm)
	log.Printf("Opening server at localhost:%s and listening for %s\n", *port, *device)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}

// Source is is
type Source struct {
	es    es.EventSource
	label string
	url   string
}

// SourceMap Maps the type (TCP, UDP, ARP, etc) to an event source
type SourceMap map[string]Source

// Add binds an event source to a packet type (ie TCP)
// creates an http handler to allow subscription to the packet type
func (sm SourceMap) Add(label string) es.EventSource {
	c := Source{
		es:    es.New(nil, nil),
		label: label,
		url:   "/pcap/" + label + "/"}
	sm[label] = c
	return c.es
}

// ServeHTTP display the available sources as json
func (sm SourceMap) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	o := make(map[string]string)
	for k := range sm {
		o[k] = sm[k].url
	}
	m, _ := json.Marshal(o)
	fmt.Fprintf(w, "%s", m)
}

func openDevice(device *string, port *string, sm SourceMap) {
	h, err := pcap.OpenLive(*device, 65535, true, 100)

	if h == nil {
		fmt.Printf("Failed to open %s : %s\n", *device, err)
	} else {
		h.SetFilter("not port " + *port)
		for pkt := h.Next(); ; pkt = h.Next() {
			if pkt != nil {
				sm.process(pkt)
			}
		}

		log.Println("timeout")
	}
}

type ports struct {
	SrcPort  uint16
	DestPort uint16
}

type addrHdr interface {
	SrcAddr() string
	DestAddr() string
	Len() int
}

func (sm SourceMap) process(pkt *pcap.Packet) {
	pkt.Decode()

	if len(pkt.Headers) >= 2 {
		if addr, ok := pkt.Headers[0].(addrHdr); ok {
			sm.route(pkt, &addr)
		}
	}
}

func (sm SourceMap) route(pkt *pcap.Packet, addr *addrHdr) {
	switch k := pkt.Headers[1].(type) {
	case *pcap.Tcphdr:
		f := &ports{k.SrcPort, k.DestPort}
		sm["tcp"].send(addr, f)
	case *pcap.Udphdr:
		f := &ports{k.SrcPort, k.DestPort}
		sm["udp"].send(addr, f)
	default:
		log.Printf("%T", k)
	}
}

func (s Source) send(addr *addrHdr, port *ports) {
	m := msg(*addr, port)
	s.es.SendEventMessage(m, "", "")
}

// Create the string representing the packet
// this is what gets sent to the browser
func msg(addr addrHdr, k *ports) string {
	return fmt.Sprintf("%s:%d %s:%d", addr.SrcAddr(), int(k.SrcPort), addr.DestAddr(), int(k.DestPort))
}
