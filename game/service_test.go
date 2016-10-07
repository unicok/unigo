package main

import (
	"fmt"
	"io"
	"math/rand"
	"testing"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	pb "lib/proto"
)

const (
	address = ":51000"
)

func TestGamePing(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithBlock(), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGameServiceClient(conn)

	md := metadata.New(map[string]string{
		"userid": fmt.Sprint(1),
	})
	ctx := metadata.NewContext(context.Background(), md)

	stream, err := c.Stream(ctx)
	if err != nil {
		t.Fatal(err)
	}

	const N = 10

	waitc := make(chan struct{})
	go func() {
		i := 0
		for {
			if i == N {
				close(waitc)
				return
			}
			in, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("replay: %v", string(in.Message))
			i++
		}
	}()

	for i := 0; i < N; i++ {
		v := r.Int31()
		t.Logf("ping with:%v", v)
		if err := stream.Send(&pb.Game_Frame{Type: pb.Game_Ping, Message: []byte(fmt.Sprint(v))}); err != nil {
			t.Fatal(err)
			return
		}
		time.Sleep(time.Second)
	}
	<-waitc
}
