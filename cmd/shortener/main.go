package main

import (
	"fmt"
	"os"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/server"
)

func main() {
	envList := os.Environ()
	// выводим первые пять элементов
	for i := 0; i < len(envList); i++ {
		fmt.Println(envList[i])
	}
	config := parseFlags()
	server.Init(config)
}
