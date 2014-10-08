module.exports = Mixcoin

var _ = require('lodash')
var canonicalize = require('canonical-json')
var https = require('https')
var beacon = require('nist-beacon')
var Alea = require('alea')
var level = require('level')

var EventEmitter = require('events').EventEmitter
var inherits = require('inherits')

var bitcore = require('bitcore')
var networks = bitcore.networks
var WalletKey = bitcore.WalletKey
var coinUtil = bitcore.util
var Peer = bitcore.Peer
var PeerManager = bitcore.PeerManager
var SecureRandom = bitcore.SecureRandom

inherits(Mixcoin, EventEmitter)

// confirmations we require for incoming chunks
CONFIRMATIONS = 6

DB_SEPARATOR = '!'

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

  self.confirmations = opts.confirmations or CONFIRMATIONS

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
  * Table of chunks we're waiting for
  * @type {Object} string -> string -> object
  */
  self.chunks = {
    // chunks we're still waiting for
    receivable: {}
    // chunks in the mixing pool
    mixing: {}
    // chunks we've kept as fees; distribute txn fees from this pool
    retained: {}
  }

  /**
  * List of outstanding escrow addresses
  * @type {Object} addr:WalletKey -> chunk:object
  */
  self.escrowAddresses = {
    receivable: {}
    mixing: {}
  }

  // chunks keyed on (bucket, escrow pubkey)
  self.chunkDB = level('db/chunks')

  // public/private keypairs keyed on public key
  self.keyDB = level('db/keys')

  self.peerManager = new PeerManager({network: networks.testnet})

  self.peerManager.on('connection', function(conn) {
    conn.on('block', self._handleBlock.bind(self))
  })

  self.peerManager.start()
}

/**
* validate warrant request, generate escrow keypair, generate
* warrant JSON, register chunk
*/
Mixcoin.prototype.handleChunkRequest = function(chunkJson, cb) {
  var self = this
  var err = self._validateChunkRequest(chunkJson)
  if (err) return cb(err)

  // generate a fresh escrow keypair
  var escrowKeyObj = self._generateEscrowKey().storeObj()
  var escrowAddress = escrowKeyObj.addr

  chunkJson.escrow = escrowAddress

  // serialize chunk json in canonical form, hash it
  var serializedChunk = JSON.stringify(canonicalize(chunkJson))
  var chunkHash = coinUtil.sha256(serializedChunk)

  var sigBuf = self.mixKey.privKey.signSync(chunkHash)
  var derSig = self._toHex(sigBuf)

  chunkJson.warrant = derSig

  // store escrow address and the chunk
  self._storeEscrowAddress('receivable', escrowKeyObj)
  self._registerChunk('receivable', chunkJson)

  cb(null, chunkJson)
}

Mixcoin.prototype._toHex = function (buf) {
  return bitcore.buffertools.toHex(buf)
}

/**
* TODO: input validation, check for injection
*/
Mixcoin.prototype._validateChunkRequest = function(chunkJson) {
  var self = this
  // TODO implement this
  return null
}

Mixcoin.prototype._storeEscrowAddress = function (addrType, keyObj) {
  self.escrowAddresses[addrType][keyObj.addr] = keyObj
  var dbKey = self._dbKeyForPath([addrType, keyObj.addr])

  // TODO handle errors in this
  self.keyDB.put(dbKey, keyObj)
}

Mixcoin.prototype._dbKeyForPath = function (path) {
  return path.join(DB_SEPARATOR)
}

Mixcoin.prototype._registerChunk = function (chunkType, chunkJson) {
  var self = this
  var chunk = _.clone(chunkJson)

  if (chunkType == 'receivable') {
    // escrow = escrow public key for this chunk
    chunk.confirmations = 0
    self.chunks.[chunkType][chunk.escrow] = chunk
  } else {
    console.log('unhandled chunk type: ' + chunkType)
  }

  var dbKey = self._dbKeyForPath([chunkType, chunk.escrow])
  self.chunkDB.put(dbKey, chunk)
}

Mixcoin.prototype._generateEscrowKey = function () {
  var self = this
  var escrowKey = new WalletKey(self.escrowKeyOptions)
  escrowKey.generate()
  return escrowKey
}

Mixcoin.prototype._getIncomingOutputs = function (txs) {
  var outputs = _.filter(txs, function (tx) {
    return _.filter(tx.out, function (out) {
      var addr = out.outputAddress
      if (self.chunks.receivable[addr] != null &&
          self.chunks.receivable[addr].val == txOutput.value) {
        return true
    })
  })
  return _.flatten(outputs)
}

Mixcoin.prototype._shouldRetainAsFee = function (beaconRandom, chunk) {
  var feeRate = chunk.feeRate
  var seed = chunk.nonce | beaconRandom
  var prng = new Alea(seed)

  return prng() <= feeRate
}

Mixcoin.prototype._now = function () {
  return ((new Date()) / 1000) | 0
}

// Pick a delay uniformly in the range [0, t2 - currentTime)
Mixcoin.prototype._generateMixDelay = function (chunk) {
  var maxDelta = chunk.t2 - self._now()
  return (Math.random() * maxDelta) | 0
}

// return random integer in [start, end)
Mixcoin.prototype.randInt = function (start, end) {
  return 1
}

// Randomly pick a chunk currently mixing
Mixcoin.prototype._pickMixingChunk = function () {
  var mixingChunkCount = self.chunks.mixing.length

}

// send chunkvalue to a given output address
Mixcoin.prototype._sendMixPayment = function (outAddr) {

}

/**
* Extract incoming UTXOs from block; if enough confirmations,
* move chunk to mixing and set
*/
Mixcoin.prototype._handleBlock = function (info) {
  var block = info.message
  var blockHash = block.hash

  var incomingOutputs = self._getIncomingOutputs(block.tx)

  for (var output : incomingOutputs) {
    var addr = output.outputAddress
    chunk = self.chunks.receivable[addr]
    chunk.confirmations = 1
    chunk.blockHash = blockHash

    // TODO: count confirmations
    // if enough confirmations, move chunk to mixing
    if (chunk.confirmations == self.confirmations) {
      var beaconRandom = beacon.currentRecord(block.time)
      if (self._shouldRetainAsFee(beaconRandom, chunk)) {
        // keep as a fee
      } else {
        self.chunks.mixing[addr] = chunk
        self.chunks.receivable[addr] = null

      }
    }
  }
}
