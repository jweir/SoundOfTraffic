(function(){
  var sot = angular.module("SoundOfTraffic",[]);


  sot.directive('navbar', function(){
    return {
      restrict: 'E',
      controller: function($scope, $http){
        $scope.sources = []

        $scope.toggle = function(s){
          return s.enabled ? s.start() : s.stop();
        }

        $http.get("/sources").success(function(d){
          angular.forEach(d, function(u,s){
            $scope.sources.push(new SOT(s,u))
          })
        })
      }
    }
  })

})();

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

  function SOT(name, url){
    this.name = name;
    this.url = url;
    this.enabled = false;
    this.sink = [];
    this.instMap = {};
    this.instCounter = 21;
    this.timer = null;
  }

  SOT.prototype.start = function(){
    console.log("starting", this.name)
    this.source = new EventSource(this.url)

    var self = this;
    this.source.onmessage = function(e){
      self.recv(e.data)
    }
    this.timer = setInterval(function(){self.drain()}, 100)
  }

  SOT.prototype.stop = function(){
    console.log("stopping", this.name)
    clearInterval(this.timer)
    this.drain();
    this.source.close();
    delete this.source;
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

  window.SOT = SOT;

  function loadMidi(){
    MIDI.loadPlugin({
      soundfontUrl: "./soundfont/",
      instrument: "acoustic_grand_piano",
      callback: function(){
        MIDI.setVolume(0,127);
        MIDI.noteOn(0,50,127, 0)
        MIDI.noteOff(0,50,1)
      }
    })
  }

  $(loadMidi);
})();
