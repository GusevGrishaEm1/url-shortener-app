package main

import (
	"context"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/config"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/server"
)

func main() {
	ctx := context.Background()
	config := config.Init()
	//go pprofSaveToFileHeapInfo()
	if err := server.StartServer(ctx, config); err != nil {
		panic(err)
	}
}

// func pprofSaveToFileHeapInfo() {
// 	time.Sleep(30 * time.Second)
// 	f, err := os.Create("profiles/result.pprof")
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer f.Close()
// 	if err := pprof.WriteHeapProfile(f); err != nil {
// 		panic(err)
// 	}
// 	println("Memory profile saved to profiles/result.pprof")
// }
