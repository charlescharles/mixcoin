var express = require('express');
var _ = require('underscore');
var bitcoin = require('bitcoinjs-lib');
var canonicalize = require('canonical-json');

var crypto = require('crypto');
var ecdsa = require('ecdsa');
var CoinKey = require('coinkey');
var sr = require('secure-random');

var app = express();

app.use(express.bodyParser());
app.use(app.router);

// do this right
var privateKey = sr.randomBuffer(32);
var ck = new CoinKey(privateKey, true);

var validateRequest = function(warrantRequestJson) {
  // check that fields have valid formats,
  // `val` is correct chunksize,
  // `return` and `out` are reasonable,
  // `fee` is a certain value, and `confirm` is reasonable
}

var generateKeyPair = function() {
  var key = bitcoin.ECKey.makeRandom();

  return {
    privateKey: key.toWIF(),
    publicKey: key.pub.getAddress().toString()
  };
}

var registerMixRequest = function(warrantRequestJson) {
  // modifies warrantRequestJson into warrantResponseJson

  var keys = generateKeyPair();
  var privateKey = keys.privateKey
  var escrowAddress = keys.publicKey

  // store request

  warrantRequestJson.escrow = escrowAddress;
  var serializedRequestJson = canonicalize.stringify(warrantRequestJson);
  // hash before signing?
  var hashed = crypto.createHash('sha256').update(serializedRequestJson).digest();
  var signature = ecdsa.sign(hashed, ck.privateKey);

  warrantRequestJson.warrant = signature;

  return warrantRequestJson;
}

app.post('/warrant', function (req, res) {
  var err = null;

  var fields = ['val', 'send', 'return', 'out', 'fee', 'nonce', 'confirm'];

  err = validateRequest(warrantRequestJson);

  if (err) {
    // validation error
  }

  warrantResponseJson = registerMixRequest(warrantRequestJson);

  if (!warrantRequestJson) {
    // dun fucked up
  }

  res.send(warrantResponseJson);
})
