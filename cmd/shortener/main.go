package main

import "github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/server"

func main() {
	err := server.StartServer(parseFlagsAndEnv())
	if err != nil {
		panic(err)
	}
}
