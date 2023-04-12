package main

import (
	"fmt"

	"github.com/ethereum-optimism/optimism/da"
)

func main() {
  client := da.NewClient("http://localhost:3000")

  fmt.Println(client.PostBatch([]byte{0x01, 0x02, 0x03}))
}
