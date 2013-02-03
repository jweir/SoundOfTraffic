package main

// Sound of Traffic (2) server

import (
	"flag"
	"fmt"
  "encoding/json"
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

type Channel struct {
  es es.EventSource
  label string
  url string
}

type ChannelList map[string]Channel

func main() {
	device := flag.String("d", "", "network device")
	flag.Parse()

	// Maps the type (TCP, UDP, ARP, etc) to an event source
	channels := make(ChannelList)

	defer createChannel("tcp", channels).Close()
	defer createChannel("udp", channels).Close()

	// TOOD "/protocol" which describes a function to
	// break the stream into sonic data

	http.Handle("/channels", channels)
	http.Handle("/", http.FileServer(http.Dir("./pub")))

	if *device == "" {
		printDevices()
	} else {
		go openDevice(device, channels)
		log.Println("Opening server at localhost:8080 and listening for ", *device)
		log.Fatal(http.ListenAndServe(":8080", nil))
	}

}

// binds an event source to a packet type (ie TCP)
// creates an http handler to allow subscription to the packet type
func createChannel(label string, channels map[string]Channel) es.EventSource {
	e := es.New(nil)
  c := Channel {
    e,
    label,
    urlFor(label),
  }
	http.Handle(urlFor(label), e)
	channels[label] = c
	return e
}

func urlFor(s string) string {
	return "/pcap/" + s + "/"
}

// display the available channels as json
func (c ChannelList) ServeHTTP(w http.ResponseWriter, r * http.Request) {
  o := make(map [string]string)
  for k, _ := range c {
    o[k] = c[k].url
  }
  m, _ := json.Marshal(o)
  fmt.Fprintf(w, "%s", m)
}

func openDevice(device *string, channels map[string]Channel) {
	h, err := pcap.OpenLive(*device, 65535, true, 100)
	if h == nil {
		fmt.Printf("Failed to open %s : %s\n", *device, err)
	} else {

		// TODO set the port via an arg
		h.SetFilter("port != 8080")
		for pkt := h.Next(); ; pkt = h.Next() {
			if pkt != nil {
				pkt.Decode()
				m, ok := prepare(pkt)
				if ok {
					channels["tcp"].es.SendMessage(m, "", "")
				}
			}
		}

		log.Println("timeout")
	}
}

// prepres the packet string and only returns TCP
// TODO return any kind of packet, not just TCP
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
