var assert = chai.assert;
var expect = chai.expect;

describe('Array', function(){

  before(function(){
    console.log(new Date());
  });

  describe('spring', function(){
    var spring;

    beforeEach(function(){
      spring = Spring.create();
    })

    afterEach(function(){
      spring.destroy();
    })

    it('winds up approaching infinite', function(){
      spring.wind(60)
    });

    it('unwinds to 0 over time', function(done){
      spring.wind(20)
      setTimeout(function(){
        expect(spring.value()).to.be.lt(20)
        done()
      }, 600)
    })
  });

  describe('genevaDrive', function(){
    it('', function(){
    });
  });
});

