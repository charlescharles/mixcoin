module.exports = Mixcoin

var _ = require('lodash')
var canonicalize = require('canonical-json')

var crypto = require('crypto')
var ecdsa = require('ecdsa')
var sr = require('secure-random')

var bitcore = require('bitcore')
var networks = bitcore.networks
var WalletKey = bitcore.WalletKey

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
}

Mixcoin.prototype.handleChunkRequest = function(chunkJson, cb {
  var self = this
  var err = self.validateChunkRequest(chunkJson)
  if (err) return cb(err)

  // generate a fresh escrow keypair
  var escrowKey = self._generateEscrowKey()
  var escrowAddress = bitcore.buffertools.toHex(escrowKey.privKey.public)

  chunkJson.escrow = escrowAddress

  // serialize chunk json in canonical form
  var canonicalizedChunk = canonicalize.stringify(chunkJson)

  // sign the serialized chunk
  var mixPrivKey = bitcore.buffertools.toHex(self.mixKey.privKey.private)

  // TODO: sign the hash instead of the chunk
  var signature = ecdsa.sign(canonicalizedChunk, mixPrivKey)

  chunkJson.warrant = signature

  // store escrow address and the chunk
  self.escrowAddresses.push(escrowKey)
  self._registerNewChunk(chunkJson)

  cb(null, chunkJson)
})

Mixcoin.prototype._registerNewChunk = function (chunkJson) {
  var chunk = _.clone(chunkJson)

  // escrow = escrow public key for this chunk
  self.chunkTable[chunk.escrow] = chunk
}

Mixcoin.prototype._generateEscrowKey = function() {
  var escrowKey = new WalletKey(self.escrowKeyOptions)
  escrowKey.generate()
  return escrowKey
  self.escrowAddresses.push(escrowKey)
}
