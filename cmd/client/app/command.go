package app

import (
	"bufio"
	"context"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	pb "github.com/ishii1648/grpc-poc/proto"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"google.golang.org/grpc"
)

const (
	defaultHost = "localhost"
	defaultPort = 10000
)

type client struct {
	name string
	host string
	port int
}

func NewClientCommand() *cobra.Command {
	client := newClient()

	cmd := &cobra.Command{
		Use:   "chat-client",
		Short: "chat client",
		Run: func(cmd *cobra.Command, args []string) {
			client.run()
		},
	}

	fs := cmd.Flags()
	client.set(fs)

	return cmd
}

func newClient() *client {
	return &client{
		host: defaultHost,
		port: defaultPort,
	}
}

func (c *client) set(fs *flag.FlagSet) {
	fs.StringVarP(&c.name, "name", "n", c.name, "from name")
	fs.StringVar(&c.host, "host", c.host, "grpc host")
	fs.IntVarP(&c.port, "port", "p", c.port, "grpc port")
}

func (c *client) run() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	addr := c.host + ":" + strconv.Itoa(c.port)
	conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	// defer conn.Close()

	client := pb.NewChatClient(conn)
	c.sendAndRecieveMsg(ctx, client)
}

func (c *client) sendAndRecieveMsg(ctx context.Context, client pb.ChatClient) error {
	stdin := bufio.NewScanner(os.Stdin)
	stream, err := client.SendAndRecieveMsg(ctx)
	if err != nil {
		log.Fatal(err)
	}
	// defer stream.CloseSend()

	stopCh := make(chan struct{})
	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				close(stopCh)
				return
			}
			if err != nil {
				log.Fatalf("Failed to receive a note : %v", err)
			}
			if in.From == c.name {
				continue
			}
			log.Printf("Got message <%s:%s>", in.From, in.Message)
		}
	}()

loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		default:
			stdin.Scan()
			text := stdin.Text()

			if err := stream.Send(&pb.SendRequest{From: c.name, Message: text}); err != nil {
				log.Fatalf("Send failed: %v", err)
			}
		}
	}

	return nil
}
