package chip8

import "flag"

// dumpDisplays prints each test ROM's final framebuffer instead of asserting
// goldens — the human-verification step when (re)recording hashes.
var dumpDisplays = flag.Bool("chip8.dump", false, "dump test ROM displays instead of asserting goldens")
