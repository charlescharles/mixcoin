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
var RpcClient = bitcore.RpcClient
var TransactionBuilder = bitcore.TransactionBuilder

inherits(Mixcoin, EventEmitter)

// confirmations we require for incoming chunks
MIN_CONFIRMATIONS = 6

// chunk size in BTC
CHUNK_SIZE = 1.0

DB_SEPARATOR = '!'

RPC_IP = '127.0.0.1'
RPC_PORT = 18333
RPC_USER = 'lightlike'
RPC_PASS = 'Thereisnone'

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

  self.rpcIp = opts.rpcIp || RPC_IP
  self.rpcPort = opts.rpcPort || RPC_PORT
  self.rpcUser = opts.rpcUser || RPC_USER
  self.rpcPass = opts.rpcPass || RPC_PASS

  self.MIN_CONFIRMATIONS = opts.confirmations || MIN_CONFIRMATIONS
  self.CHUNK_SIZE = opts.chunkSize || CHUNK_SIZE

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
    pool: {}
    // chunks we've kept as fees; distribute txn fees from this pool
    retained: {}
  }

  /**
  * List of outstanding escrow addresses
  * @type {Object} addr:WalletKey -> chunk:object
  */
  self.escrowKeys = {
    receivable: {}
    pool: {}
  }

  // chunks keyed on (bucket, escrow pubkey)
  self.chunkDB = level('db/chunks')

  // public/private keypairs keyed on public key
  self.keyDB = level('db/keys')

  self.rpc = new RpcClient({
    protocol: 'http',
    user: self.rpcUser,
    pass: self.rpcPass,
    host: self.rpcIp,
    port: self.rpcPort
  })

  self.peerManager = new PeerManager({network: networks.testnet})
  self.peerManager.addPeer(new Peer(self.rpcIp, self.rpcPort))

  self.peerManager.on('connection', function(conn) {
    conn.on('block', self.emit('block'))
  })

  self.peerManager.start()

  // events
  self.on('block', self._checkForReceivedTxns)
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
  self._registerEscrowKey('receivable', escrowKeyObj)
  self._registerNewChunk('receivable', chunkJson)

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

// add escrow key to db
// import address to bitcoind
Mixcoin.prototype._registerEscrowKey = function (addrType, keyObj) {
  var self = this
  self.escrowAddresses[addrType][keyObj.addr] = keyObj
  var dbKey = self._dbKeyForPath([addrType, keyObj.addr])

  // TODO handle errors in this
  self.keyDB.put(dbKey, keyObj)

  // import address to bitcoind
  // TODO import address or importprivkey?
  self.rpc.importAddress(keyObj.addr, '', true, function (err, res) {
    // handle an error
  })
}

Mixcoin.prototype._dbKeyForPath = function (path) {
  var self = this
  return path.join(DB_SEPARATOR)
}

Mixcoin.prototype._registerNewChunk = function (chunkType, chunkJson) {
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
  var self = this
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
  var self = this
  var feeRate = chunk.feeRate
  var seed = chunk.nonce | beaconRandom
  var prng = new Alea(seed)

  return prng() <= feeRate
}

// return current time in ms
Mixcoin.prototype._now = function () {
  return (new Date()).getTime()
}

// Pick a delay in ms uniformly in the range [0, t2 - currentTime)
Mixcoin.prototype._generateMixDelay = function (chunk) {
  var self = this
  var maxDelta = chunk.t2 * 1000 - self._now()
  return (Math.random() * maxDelta) | 0
}

// return random integer in [start, end)
Mixcoin.prototype.randInt = function (start, end) {
  return 1
}

// Return a random chunk currently in the mixing pool
Mixcoin.prototype._randomMixingChunk = function () {
  var self = this
  var mixingChunkCount = self.chunks.mixing.length
  // TODO finish this
}

// send chunkvalue to a given output address
Mixcoin.prototype._sendMixPayment = function (outAddr) {
  var self = this
  var chunk = self._randomMixingChunk()
  var fromAddr = chunk.escrow

  // TODO make sure peermanager is ready first

  var unspent = [{
    txid: chunk.utxoInfo.txid,
    vout: chunk.utxoInfo.vout,
    address: fromAddr,
    scriptPubKey: chunk.utxoInfo.scriptPubKey,
    amount: self.CHUNK_SIZE,
    confirmations: chunk.utxoInfo.confirmations
  }]

  // TODO handle txn fees
  var out = [{
    address: outAddr,
    amount: self.CHUNK_SIZE
  }]

  var privKey = self.escrowKeys.pool[fromAddr].priv
  var txOpts = {}
  var outTx = new TransactionBuilder(txOpts)
                    .setUnspent(unspent)
                    .setOutputs(out)
                    .sign([privKey])
                    .build()

  // TODO log the output tx
  // TODO set active connection on self
  self.peerman.getActiveConnection()sendTx(outTx)
}

Mixcoin.prototype._beginMixingChunk = function (addr) {
  var self = this
  self._addToPool(addr)

  var chunk = self.chunks.pool[addr]
  var outAddr = chunk.outAddr
  var delay = self._generateMixDelay(chunk)
  var sendPayment = function () {self._sendMixPayment(outAddr)}

  setTimeout(sendPayment.bind(self), delay)
}

// add the chunk corresponding to this escrow address to the mixing pool,
// remove from receivable chunks table
Mixcoin.prototype._addToPool = function (escrowAddr) {
  var self = this
  self.chunks.pool[escrowAddr] = self.chunks.receivable[escrowAddr]
  delete self.chunks.receivable[escrowAddr]

  // TODO delete from db and re-add in new namespace
}

//
Mixcoin.prototype._checkForReceivedTxns = function () {
  var self = this
  var rpc = self.rpc

  rpc.listReceivedByAddress(self.MIN_CONFIRMATIONS, false, function (err, receivedData) {
    var received = receivedData.result
    var receivedAddrs = []

    // get txids
    for (var account : received) {
      var addr = account.addr
      if (self.chunks.receivable[addr] && account.amount >= self.CHUNK_SIZE) {
        // TODO Handle multiple txids
        receivedAddrs.push([addr, account.txids[0]])
      }
    }

    // get vouts and scriptPubKeys
    for (var addrTx : receivedAddrs) {
      var addr = addrTx[0]
      var txid = addrTx[1]

      // closure over addr and txid
      (function (addr, txid) {
        rpc.getTransaction(txid, function (err, txData) {
          var tx = txData.result
          self.chunks.receivable[addr].utxoInfo = {}

          var utxoInfo = self.chunks.receivable[addr].utxoInfo

          // TODO make sure txn.amount >= self.CHUNK_SIZE
          // if it's over, remove some
          // also make sure that context is accessible here
          // TODO handle multiple details
          // _.assign(self.chunks.receivable[addr], {
          //   txid: txid
          //   vout: tx.details[0].vout
          //   blocktime: tx.blocktime
          //   blockHash: tx.blockHash
          // })
          utxoInfo.txid = txid
          utxoInfo.vout = tx.details[0].vout
          utxoInfo.blocktime = tx.blocktime
          utxoInfo.blockHash = tx.blockHash

          rpc.getTxOut(txid, utxoInfo.vout, true, function (err, txOutData) {
            var txOut = txOutData.result
            utxoInfo.scriptPubKey = txOut.scriptPubKey.hex
            utxoInfo.confirmations = txOut.confirmations

            self._beginMixingChunk(addr)
          })
        })
      })(addr, txid)
    }
  })
}
