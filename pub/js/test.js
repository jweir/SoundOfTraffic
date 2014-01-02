// Window the spring up
// it will unwind over time
function Spring(t){
  this.wound = 0;
  var self = this;
  self.clock = setInterval(function(){ self.unwind(); }, t || 100)
}

Spring.create = function(){
  return new Spring();
}

Spring.prototype.wind = function(v){
  this.wound += v;
  return this;
}

Spring.prototype.unwind = function(){
  // TODO function to unwind with greater speed as the wound value
  // increases
  this.wound *= 0.75;
  console.log(this.wound)
  return this;
}

Spring.prototype.value = function(){
  return this.wound;
}

Spring.prototype.destroy = function(){
  clearInterval(this.clock);
  delete this;
}
