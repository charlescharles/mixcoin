var run = function() {
  var bitcore = require('bitcore');
  var NetworkMonitor = bitcore.NetworkMonitor;

  var config = {
    networkName: 'testnet',
    host: '127.0.0.1',
    port: 18333
  };

  var nm = new NetworkMonitor.create(config);
  nm.incoming('mpKLjUc3cMRk2JHDdsFqajoxNpA6QURakn', function(tx) {
    debugger
  });

  // connect to bitcoin network and start listening
  nm.start();

};

module.exports.run = run;
if (require.main === module) {
  run();
}
