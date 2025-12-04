package main

const KSIZE = 20
const ALPHA = 3
const REPLACEMENT_FACTOR = 5
const NODE_ID_BUFFER_SIZE = 32 // 20 bytes in 160-bit node ID, but we are using sha-256 so change to 32 bytes