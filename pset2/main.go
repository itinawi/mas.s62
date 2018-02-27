package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
	"log"
)

// A hash is a sha256 hash, as in pset01
type Hash [32]byte

// ToString gives you a hex string of the hash
func (self Hash) ToString() string {
	return fmt.Sprintf("%x", self)
}

// Blocks are what make the chain in this pset; different than just a 32 byte array
// from last time.  Has a previous block hash, a name and a nonce.
type Block struct {
	PrevHash Hash
	Name     string
	Nonce    string
}

// ToString turns a block into an ascii string which can be sent over the
// network or printed to the screen.
func (self Block) ToString() string {
	return fmt.Sprintf("%x %s %s", self.PrevHash, self.Name, self.Nonce)
}

// Hash returns the sha256 hash of the block.  Hopefully starts with zeros!
func (self Block) Hash() Hash {
	return sha256.Sum256([]byte(self.ToString()))
}

// BlockFromString takes in a string and converts it to a block, if possible
func BlockFromString(s string) (Block, error) {
	var bl Block

	// check string length
	if len(s) < 66 || len(s) > 100 {
		return bl, fmt.Errorf("Invalid string length %d, expect 66 to 100", len(s))
	}
	// split into 3 substrings via spaces
	subStrings := strings.Split(s, " ")

	if len(subStrings) != 3 {
		return bl, fmt.Errorf("got %d elements, expect 3", len(subStrings))
	}

	hashbytes, err := hex.DecodeString(subStrings[0])
	if err != nil {
		return bl, err
	}
	if len(hashbytes) != 32 {
		return bl, fmt.Errorf("got %d byte hash, expect 32", len(hashbytes))
	}

	copy(bl.PrevHash[:], hashbytes)

	bl.Name = subStrings[1]

	// remove trailing newline if there; the blocks don't include newlines, but
	// when transmitted over TCP there's a newline to signal end of block
	bl.Nonce = strings.TrimSpace(subStrings[2])

	// TODO add more checks on name/nonce ...?

	return bl, nil
}

func main() {

	fmt.Printf("NameChain Miner v0.1\n")

	/*
	// This checks that validate works!!
	tests := []string{
		"0000000006e13067fb12f4faff41756fa2f8c4e5d80f03edff8641ba0c82e411 fortenforge 100d87819",
		"000000000b87b8464b968ca1a161feacde4233c555608be9d142983dc9670ef4 bsnowden3 1042113315",
		"0000000016ac2685a2c8177082c903c318e7e44d2aa1018270cc5dae81eb7af7 U+1F4A9 25828249403",
		"000000004bf42a26b8bc5e464dd3c489df668e98094c2f78da72e9187590461a bsnowden3 w5340490804",
		"00000000311cdc27904de495f5c5566688f18afed76ad5888d24273a937c1744 bsnowden3 1025129623",
		"00000000391e550fdcc98c7097e9a052c51d5bda8f2fced3d7b1c54db82809ac U+1F4A9 22586278263",
		"000000002d881a00da989992770af5a45f78719dc0b4aa8119cdc285e791a286 U+1F4A9 34154720403",
		"000000000a5710659e075cd608cad37796dee28da94198a0c890b5ba64768b5c bsnowden3 1185296952",
		"000000001a8baf4c76afc61ea407fb04a8a5b5aa43285535a9088ed6aa0254e3 U+1F4A9 4425056304",
		"00000000468a55bb139c8e846d73807227d409b2f60dff97544f793928739527 U+1F4A9 12869216422",
		"0000000006e13067fb12f4faff41756fa2f8c4e5d80f03edff8641ba0c82e411 bad 100d87819",
	}
	for i, test := range tests {
		test_block, err := BlockFromString(test)

		log.Println(CheckWork(test_block, 33))
		if err != nil {
			log.Println("error")
			log.Println(i)
		}
	}
	*/
	// Your code here!

	// Basic idea:
	// Get tip from server, mine a block pointing to that tip,
	// then submit to server.
	// To reduce stales, poll the server every so often and update the
	// tip you're mining off of if it has changed.

	for true {
		// we want to keep mining until program is killed by user.
		log.Println("About to start a new main instance")
		original_block, err := GetTipFromServer()
		curr_block := original_block
		if err != nil {
			log.Println("Can't get first ever tip from server.")
		}

		// a channel to tell it to stop
		stopchan := make(chan struct{})
		// a channel to signal that it's stopped
		stoppedchan:= make(chan struct{})

		go Mine(original_block, 33, stoppedchan, stopchan)
		for curr_block == original_block {
			// while this is true do not interrupt mining
			curr_block, err = GetTipFromServer()
			if err != nil {
				log.Println("Can't get new tip from server. Is it down?")
			}
			// ping the server every 60 seconds.
			time.Sleep(60*time.Second)
		}
		// that means the current block is changed so tell the miners to stop
		close(stopchan)
		<-stoppedchan
		fmt.Printf("Stopped mining. New block detected. Old block was: "+ original_block.ToString())
	}

	return
}
