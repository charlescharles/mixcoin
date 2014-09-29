module.exports = Mixcoin

var _ = require('lodash')
var canonicalize = require('canonical-json')

var crypto = require('crypto')
var ecdsa = require('ecdsa')
var sr = require('secure-random')

var EventEmitter = require('events').EventEmitter

var bitcore = require('bitcore')
var networks = bitcore.networks
var WalletKey = bitcore.WalletKey
var coinUtil = bitcore.util

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

  var signature = ecdsa.sign(chunkHash, self.mixKey.privKey.private)
  debugger
  chunkJson.warrant = signature

  // store escrow address and the chunk
  self.escrowAddresses.push(escrowKey)
  self._registerNewChunk(chunkJson)

  cb(null, chunkJson)
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

Mixcoin.prototype._generateEscrowKey = function() {
  var self = this
  var escrowKey = new WalletKey(self.escrowKeyOptions)
  escrowKey.generate()
  return escrowKey
  self.escrowAddresses.push(escrowKey)
}
