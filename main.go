package main

import (
	"fmt"
	"lan-file/receive"
	"lan-file/send"
	"os"
)

func main() {
	list := os.Args
	if list[1] == "s" {
		fmt.Println("send file...")
		send.SendFile(list[2], list[3], list[4])
	} else if list[1] == "r" {
		fmt.Println("receiving file...")
		receive.ReceiveFile(list[2])
	}

}
