(function(){

  // Splits a string of address info into an object
  function split(addr){
    var sd = addr.split(/ /), s = sd[0].split(/:/), d = sd[1].split(/:/);

    return {
      sourceIP: s[0],
      sourcePort: s[1],
      destIP: d[0],
      destPort: d[1]
    }
  }

  function SOT(){
    this.sink = [];
    this.instMap = {};
    this.instCounter = 21;
    debugger
  }

  SOT.prototype.recv = function recv(message){
    console.log("recv")
    this.sink.push(message);
  }

  SOT.prototype.drain = function(){
    console.log("drain")
    this.sonify(this.portsToInst(this.sink));
    this.sink = [];
  }

  SOT.prototype.sonify = function(map){
    var self = this;
    _.each(map, function(arr, port){
      var note = self.instMap[port] = self.instMap[port] || self.instCounter ++;
      MIDI.noteOn(0,note, arr.length);
      MIDI.noteOff(0,note, 1);
    })
  }

  // Format is IP:port IP:port
  SOT.prototype.portsToInst = function(sink){
    var map = {}
    _.map(sink, function(d){
      addr = split(d)
      map[addr.sourcePort] = map[addr.sourcePort] || []
      map[addr.sourcePort].push(addr.sourceIP)

      map[addr.destPort] = map[addr.destPort] || []
      map[addr.destPort].push(addr.destIP)
    })

    return map;
  }

  function listen(){
    var source = new EventSource('/pcap/tcp/');
    var sot = new SOT();

    source.onmessage = function(e){
      sot.recv(e.data)
    }

    return sot;
  }

  function start(){
    var sot = listen();
    setInterval(function(){sot.drain()}, 100)
  }

  $(listen);

  function loadMidi(){
    MIDI.loadPlugin({
      soundfontUrl: "./soundfont/",
      instrument: "acoustic_grand_piano",
      callback: function(){
        MIDI.setVolume(0,127);
        MIDI.noteOn(0,50,127, 0)
        MIDI.noteOff(0,50,1)
        start()
      }
    })
  }

  $(loadMidi);
}())
