package main

const KSIZE = 2
const ALPHA = 3
const BSIZE = 5
const REPLACEMENT_FACTOR = 5
const NODE_ID_BUFFER_SIZE = 32 // 20 bytes in 160-bit node ID, but we are using sha-256 so change to 32 bytes
const NODE_ID_BIT_SIZE = 32 * 8
const STOR_REPLICATION = 5 // how many nodes to replicate a key/value to store