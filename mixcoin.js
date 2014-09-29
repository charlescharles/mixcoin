module.exports = Mixcoin

var _ = require('lodash')
var canonicalize = require('canonical-json')

var EventEmitter = require('events').EventEmitter

var bitcore = require('bitcore')
var networks = bitcore.networks
var WalletKey = bitcore.WalletKey
var coinUtil = bitcore.util
var Peer = bitcore.Peer
var PeerManager = bitcore.PeerManager


/**
* An implementation of the Mixcoin accountable mixing service protocol
* @param {string|Buffer} opts
*/
function Mixcoin (opts) {
  var self = this

  if (!(self instanceof Mixcoin)) return new Mixcoin(opts)
  EventEmitter.call(self)

  if (!opts.privateKey) return new Error('you must supply a private key')

  self.ready = false
  self.listening = false
  self._binding = false
  self._destroyed = false

  self.rpcIp = opts.rpcIp
  self.rpcPort = opts.rpcPort

  // generate public key
  var keyOptions = {
    network: networks.testnet
  }

  self.mixKey = new WalletKey(keyOptions)
  self.mixKey.fromObj({
    priv: opts.privateKey
  })

  self.escrowKeyOptions = {
    network: networks.testnet
  }

  /**
  * Table of outstanding chunks
  * @type {Object} addr:string -> chunk object
  */
  self.chunkTable = {}

  /**
  * List of outstanding escrow addresses
  * @type {WalletKey|Buffer}
  */
  self.escrowAddresses = []

  self.peerManager = new PeerManager({network: networks.testnet})

  self.peerManager.on('connection', function(conn) {
    conn.on('block', self._handleBlock.bind(self))
  })

  self.peerManager.start()
}

Mixcoin.prototype.handleChunkRequest = function(chunkJson, cb) {
  var self = this
  var err = self._validateChunkRequest(chunkJson)
  if (err) return cb(err)

  // generate a fresh escrow keypair
  var escrowKey = self._generateEscrowKey()
  var escrowAddress = bitcore.buffertools.toHex(escrowKey.privKey.public)

  chunkJson.escrow = escrowAddress

  // serialize chunk json in canonical form, hash it
  var serializedChunk = JSON.stringify(canonicalize(chunkJson))
  var chunkHash = coinUtil.sha256(serializedChunk)

  var sigBuf = self.mixKey.privKey.signSync(chunkHash)
  var derSig = self._toHex(sigBuf)

  debugger
  chunkJson.warrant = derSig

  // store escrow address and the chunk
  self.escrowAddresses.push(escrowKey)
  self._registerNewChunk(chunkJson)

  cb(null, chunkJson)
}

Mixcoin.prototype._toHex = function (buf) {
  return bitcore.buffertools.toHex(buf)
}

Mixcoin.prototype._validateChunkRequest = function(chunkJson) {
  var self = this
  // TODO implement this
  return null
}

Mixcoin.prototype._registerNewChunk = function (chunkJson) {
  var self = this
  var chunk = _.clone(chunkJson)

  // escrow = escrow public key for this chunk
  self.chunkTable[chunk.escrow] = chunk
}

Mixcoin.prototype._generateEscrowKey = function () {
  var self = this
  var escrowKey = new WalletKey(self.escrowKeyOptions)
  escrowKey.generate()
  return escrowKey
}

Mixcoin.prototype._handleBlock = function (info) {
  var block = info.message
  var txs = block.tx

  for (var tx in txs) {
    var txsIn = tx.in
    var txsOut = tx.out

    // check if there's a transaction sending correct value to
    // one of the escrow addresses
    // TODO assume all chunk values are the same?
  }
}
