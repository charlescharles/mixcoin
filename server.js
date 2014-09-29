var _ = require('underscore')
var mixcoin = require('mixcoin')
var http = require('http')
var sr = require('secure-random')

var options = {}
options.privateKey =sr.randomBuffer(32)

var mixcoinServer = mixcoin.Mixcoin(options)

var handleChunkRequest = function (req, res) {
  var chunk = JSON.parse(req.body)
  console.log(this)

  responseJson = mixcoinServer.handleChunkRequest(chunk)

  res.writeHead(200, {'Content-Type': 'application/json'})
  res.write(JSON.stringify(responseJson))
  res.end()
}

var server = http.createServer(handleChunkRequest)

server.listen(18433, '127.0.0.1')

console.log('server listening on 127.0.0.1:18433')
