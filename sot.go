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
	http.Handle("/log/", es)
	http.HandleFunc("/", index)

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

    h.Setfilter("port != 8080")
		for pkt := h.Next(); ; pkt = h.Next() {
			if pkt != nil {
				pkt.Decode()
        es.SendMessage(pkt.String(),"","")
				// TODO get the tcp data and port
				// if len(pkt.Headers) >= 2 {
					// if addr, ok := pkt.Headers[0].(addrHdr); ok {
						// es.SendMessage(addr.DestAddr()+" < "+addr.SrcAddr(), "", "")
					// }
				// }
			}
		}

		log.Println("timeout")
	}

}

func printDevices() {
	devs, err := pcap.Findalldevs()

	if len(devs) == 0 {
		fmt.Printf("Error: no devices found. %s\n", err)
	} else {
		fmt.Println("Available network devices")
		for _, dev := range devs {
			fmt.Printf("  %s \n", dev.Name)
		}
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, template())
}

func template() string {
	return `
  <!DOCTYPE html>

  <html>
  <head>
  <title>ESTail</title>
  <link href='http://fonts.googleapis.com/css?family=Droid+Sans+Mono' rel='stylesheet' type='text/css'>
  <script src="//cdnjs.cloudflare.com/ajax/libs/jquery/1.9.0/jquery.min.js"></script>
  <style type="text/css">
  body {
    margin: 20px;
    font: 12px/18px 'Droid Sans Mono', san-serif;
    color: #CCCCAA;
    background: #222;
  }

  hr {
    border:0;
    height: 1px;
    line-height: 1px;
    border-top: 1px #590 solid;
    margin: 9px 0 10px;
  }

  </style>
  </head>

  <body>
  <h4>ESTail</h4>
  <div id="log"></div>

  <script>
  (function(){

    // clean up some shell color codes
    function parse(s){
      return s.replace(/\[\d{1,}m/img,"")
    }

    function listen(channel, target){
      var source = new EventSource('/log/');
      var addMark = false;

      setInterval(function(){
        if(addMark){ target.append($("<hr/>")); addMark = false}
      },3000)

      source.onmessage = function(e){
        addMark = true;
        var d = $("<div/>")
        target.append(d.text(parse(e.data)))
        $(document).scrollTop($(document).height())
      }
    }

    listen("log", $("#log"))
  })()
  </script>
  </body>

  </html>
  `
}
