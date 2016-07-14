package main

import (
	"testing"

	"golang.org/x/net/context"

	pb "github.com/amorwilliams/bodoni/snowflake/proto"
	"google.golang.org/grpc"
)

const (
	address = "localhost:50003"
	testKey = "test_key"
)

func TestCasDelay(t *testing.T) {
	casDelay()
}

func TestSnowflake(t *testing.T) {
	// set up a connection to the server
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		t.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewSnowflakeServiceClient(conn)

	// contact the server and print out its response.
	r, err := c.Next(context.Background(), &pb.Snowflake_Key{Name: testKey})
	if err != nil {
		t.Fatalf("could not get next value: %v", err)
	}
	t.Log(r.Value)
}

func BenchmarkSnowflake(b *testing.B) {
	// set up a connection to the server
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		b.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewSnowflakeServiceClient(conn)

	for i := 0; i < b.N; i++ {
		// contact the server and print out its response.
		_, err := c.Next(context.Background(), &pb.Snowflake_Key{Name: testKey})
		if err != nil {
			b.Fatalf("could not get next value %v", err)
		}
	}
}

func TestSnowflakeUUID(t *testing.T) {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		t.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewSnowflakeServiceClient(conn)

	// Contact the server and print out its response.
	r, err := c.GetUUID(context.Background(), &pb.Snowflake_NullRequest{})
	if err != nil {
		t.Fatalf("could not get next value: %v", err)
	}
	t.Logf("%b", r.Uuid)
}

func BenchmarkSnowflakeUUID(b *testing.B) {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		b.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewSnowflakeServiceClient(conn)

	for i := 0; i < b.N; i++ {
		// Contact the server and print out its response.
		_, err := c.GetUUID(context.Background(), &pb.Snowflake_NullRequest{})
		if err != nil {
			b.Fatalf("could not get uuid: %v", err)
		}
	}
}
