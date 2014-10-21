# Mixcoin

This is an implementation of the Mixcoin protocol (Bonneau, Narayanan, Miller, Clark, Kroll, Felten 2014). [link to paper](https://eprint.iacr.org/2014/077.pdf)

# How to run it

Download golang

`go install btcd`
`go install btcwallet`

`btcd -u username -P password --simnet`
`btcwallet -u username -P password --rpccert ~/path/to/cert.crt --rpckey ~/path/to/key/csr --simnet`
`go run api_server.go`

# Documentation

## Server

### API

```json
{
  "val": 100000,
  "send": 1411790000,
  "return": 1411800000,
  "out": "3J98t1WpEZ73CNmQviecrnyiWrnqRhWNLy",
  "fee": 12,
  "nonce": 857349958,
  "confirm": 6
}
```

returns the following, if accepted:

```json
{
  "val": 100000,
  "send": 1411790000,
  "return": 1411800000,
  "out": "3J98t1WpEZ73CNmQviecrnyiWrnqRhWNLy",
  "fee": 12,
  "nonce": 857349958,
  "confirm": 6,,
  "escrow": "3J98t1WpEZ73CNmQviecrnyiWrnqRhWNLy",
  "warrant": "3J98t1WpEZ73CNmQviecrnyiWrnqRhWNLy"
}
```


`val`: value to be mixed, in satoshi

`send`: deadline for sending funds, in unix epoch time

`return`: deadline for distributing funds, in unix epoch time

`out`: transfer destination address

`fee`: mixing fee, in basis points

`nonce`: random 32 bit number

`confirm`: number of blocks mix will wait for payment confirmation

`escrow`: escrow address to send `val` to

`warrant`: request JSON, along with `escrow` field, canonicalized, serialized, and signed with mix's private key
