package main

import (
	"fmt"
	"os"
	"prsync/parallelRsync"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Printf("Usage: %s sourcePath destinationPath\n", os.Args[0])
		fmt.Printf("Example: %s /vol8/home /vol9\n", os.Args[0])
		return
	}
	source := os.Args[1]
	destination := os.Args[2]
	err := parallelRsync.PRsyncStart(source, destination)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("\nprsync complate!\n")
	}
}
