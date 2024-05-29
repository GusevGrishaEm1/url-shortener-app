package grpc

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/models"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func ExampleStartServer() {
	cmd := exec.Command("go", "run", "./start_server/example_start_server.go")
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	md := metadata.New(map[string]string{
		string(models.UserID): "1",
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	conn, err := grpc.DialContext(
		ctx,
		":50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithIdleTimeout(5*time.Second),
	)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	client := NewShortenerServiceClient(conn)
	request := &CreateShortURLRequest{URL: "http://example.com"}

	responseCreate, err := client.CreateShortURL(ctx, request)
	if err != nil {
		panic(err)
	}

	responseGet, err := client.GetByShortURL(ctx, &GetByShortURLRequest{ShortURL: "abc123"})
	if err != nil {
		panic(err)
	}

	fmt.Println(responseCreate.URL)
	fmt.Println(responseGet.OriginalURL)
	// Output:http://short.url/abc123
	// http://example.com
}
