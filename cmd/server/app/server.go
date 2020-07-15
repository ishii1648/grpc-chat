package app

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	pb "github.com/ishii1648/grpc-poc/proto"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"google.golang.org/grpc"
)

const (
	defaultHost = "localhost"
	defaultPort = 10000
)

type server struct {
	host    string
	port    int
	stopCh  chan struct{}
	streams []pb.Chat_SendAndRecieveMsgServer
	mu      sync.Mutex
	from    string
}

type chat struct {
	from    string
	message string
}

func NewServerCommand() *cobra.Command {
	server := newServer()

	cmd := &cobra.Command{
		Use:   "chat-server",
		Short: "chat server",
		Run: func(cmd *cobra.Command, args []string) {
			server.run()
		},
	}

	fs := cmd.Flags()
	server.set(fs)

	return cmd
}

func newServer() *server {
	return &server{
		host:   defaultHost,
		port:   defaultPort,
		stopCh: make(chan struct{}),
	}
}

func (s *server) set(fs *flag.FlagSet) {
	fs.StringVar(&s.host, "host", s.host, "grpc host")
	fs.IntVarP(&s.port, "port", "p", s.port, "grpc port")
}

func (s *server) run() {
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.host, s.port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterChatServer(grpcServer, newServer())
	grpcServer.Serve(lis)
}

func (s *server) SendAndRecieveMsg(stream pb.Chat_SendAndRecieveMsgServer) error {
	log.Print("connect from client")
	s.streams = append(s.streams, stream)
	chatCh := make(chan *chat)

	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				close(s.stopCh)
				return
			}
			if err != nil {
				close(s.stopCh)
				return
			}
			if in.Message == "exit" {
				log.Print("recieve exit message from client")
				return
			}

			chatCh <- &chat{
				from:    in.From,
				message: in.Message,
			}
		}
	}()

loop:
	for {
		select {
		case <-s.stopCh:
			log.Print("break stream loop")
			break loop
		case chat := <-chatCh:
			log.Printf("recieve chat message <%s:%s>", chat.from, chat.message)
			for _, stream := range s.streams {
				stream.Send(&pb.SendResult{
					From:    chat.from,
					Message: chat.message,
				})
			}
		}
	}

	return nil

}
