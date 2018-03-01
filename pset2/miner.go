package main

import(
  "runtime"
  "log"
  "fmt"
  //"math/big"
)
// This file is for the mining code.
// Note that "targetBits" for this assignment, at least initially, is 33.
// This could change during the assignment duration!  I will post if it does.

// Mine mines a block by varying the nonce until the hash has targetBits 0s in
// the beginning.  Could take forever if targetBits is too high.
// Modifies a block in place by using a pointer receiver.
func Mine(original_tip Block, targetBits uint8, stoppedchan chan struct{}, stopchan chan struct{}) {
	// your mining code here
	// also feel free to get rid of this method entirely if you want to
	// organize things a different way; this is just a suggestion

  // close the stoppedchan when this func
  // exits
  defer close(stoppedchan)
  // setup work
  cpus := runtime.NumCPU()
  fmt.Printf("This PC has %d CPUs\n", cpus)
  prev_hash := original_tip.Hash()
  my_name := "Ihssanremote"
  my_nonce := "-1"
  new_block_string := fmt.Sprintf("%x %s %s", prev_hash, my_name, my_nonce)
  running := true
  for j := 0; j < cpus; j++ {
    cpuBlock, err := BlockFromString(new_block_string)
    if err != nil {
      log.Println("encountered error getting block from string")
    }
    go cpuMine(cpuBlock, targetBits, j, &running)
  }

  defer func(){
    // TODO: do teardown work
    log.Println("Exiting... (teardown work)")
    running = false
  }()
  for {
    select {
      default:
        // the work that will be interrupted when the tip block is updated
        // TODO make sure that all the processes you spawn are indeed
        //  interrupted when the stop command is received.

      case <-stopchan:
        // stop
        log.Println("Got an order to stop.")
        return
    }
  }

	return
}

// CheckWork checks if there's enough work
func CheckWork(bl Block, targetBits uint8) bool {
	// your checkwork code here
	// feel free to inline this or do something else.  I just did it this way
	// so I'm giving empty functions here.
  hash := bl.Hash()
  zero_byte := byte(0)
  for i, hashbyte := range hash {
    // another way to do this is to mask the first 64 bits and do a comparison
    //  with target bits.
    if i <= 3 {
      if hashbyte > zero_byte {
        return false
      }
    } else if i == 4 {
      mask := byte(1<<7)
      if hashbyte&mask == mask{
        return false
      }
    } else {
      // we only need first 33 bits
      break
    }
  }
  return true
}

func cpuMine(cpuBlock Block, targetBits uint8, cpuId int, running *bool) {
  for *running {
    count := 300000000
    cpuBlock.Nonce = "0"
    printMod := 100000
    for !CheckWork(cpuBlock, targetBits) {
      cpuBlock.Nonce = fmt.Sprintf("%d+%d", count, cpuId)
      if (count % printMod == 0) {
        fmt.Printf("Did %d tries from cpu: %d\n", count, cpuId)
        fmt.Printf("FYI, the previous block hash is %x\n", cpuBlock.PrevHash)
      }
      count += 1
    }
    log.Println("We have a found a block that works!!!")
    fmt.Printf("Our block was: " + cpuBlock.ToString() + "\n")
    err := SendBlockToServer(cpuBlock)
    if err != nil {
      log.Println("\n\n\ngot an error when trying to send block to server.\n\n\n")
    }
  }
}
