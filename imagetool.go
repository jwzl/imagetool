package main

import (
	"fmt"
	"github.com/jwzl/imagetool/image"
)


func main() {
	err := image.GenerateRKImage("update.img", "./package-file", "./Image/parameter.txt")
	if err != nil {
		fmt.Println("Error", err)	
		return
	}

	fmt.Println("[Done]")
}
