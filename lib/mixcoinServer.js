var mixcoin = require('./mixcoin')
var http = require('http')

var sr = require('secure-random')
var bitcore = require('bitcore')

var options = {}
options.privateKey = bitcore.buffertools.toHex(sr.randomBuffer(32))

var mixcoinServer = mixcoin(options)

var handleChunkRequest = function (req, res) {
  if (req.method == 'POST') {
    var body = ''
    req.on('data', function(data) {
      body += data

      if (body.length > 1e6) {
        req.connection.destroy()
      }
    })
    req.on('end', function() {
      var chunk = JSON.parse(body)

      mixcoinServer.handleChunkRequest(chunk, function (err, responseJson) {
        res.writeHead(200, {'Content-Type': 'application/json'})
        res.write(JSON.stringify(responseJson))
        res.end()
      })

    })
  }
}

var server = http.createServer(handleChunkRequest)

server.listen(18433, '127.0.0.1')

console.log('server listening on 127.0.0.1:18433')
