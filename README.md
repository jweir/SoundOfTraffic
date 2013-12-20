# Sound of Traffic _2_

The server side archictecture is based on one or more Sensors and a single Observer.

A Sensor takes a stream of data, filters it and passes the stream to an
Observer. A Sensor connects to the Observer via a tdb protocol. A sensor can
run on the same, or difference, server than the Observer.

An Observer passes Sensor streams down to the Client via EventSource channels.
Each Sensor has its own channel. The Observer also handls client
authentication.

The Client subscribes to one or more Sensor channels.  The Client then
visualizes and sonifies the data stream as defined by the script.

Example

  A TCPCap Sensor takes all network traffic and filters it into the follow format

  PORT IPADDRESS-SRC IPADDRESS-DEST PACKEST-SIZE REGION-SRC REGION-DEST
  _The region might be country or contintent, etc._

  The Sensor is configured to connect to a Observer running o.example.com with an AUTHKEY.

  The data is streamed and encrypted to the Observer.

  The client script recieves the stream and assigns sounds loops based on region and direction.

    go build sot.go; sudo ./sot -i en1 
    # it needs to run as root for PCAP to have access to the packets
    # -i is the interface, just run sudo ./sot for a list of interfaces

see http://www.geek.com/apps/this-is-what-a-ddos-attack-looks-like-1552975/ for some visuals inspiration
