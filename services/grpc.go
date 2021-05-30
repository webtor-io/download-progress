package services

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	pb "github.com/webtor-io/download-progress/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	grpcHostFlag = "grpc-host"
	grpcPortFlag = "grpc-port"
)

func RegisterGRPCFlags(f []cli.Flag) []cli.Flag {
	return append(f,
		cli.StringFlag{
			Name:   grpcHostFlag,
			Usage:  "grpc listening host",
			Value:  "",
			EnvVar: "GRPC_HOST",
		},
		cli.IntFlag{
			Name:   grpcPortFlag,
			Usage:  "grpc listening port",
			Value:  50051,
			EnvVar: "GRPC_PORT",
		},
	)
}

type GRPC struct {
	pb.UnimplementedDownloadProgressServer
	host string
	port int
	ln   net.Listener
	sp   *StatPool
}

func NewGRPC(c *cli.Context, sp *StatPool) *GRPC {
	return &GRPC{
		host: c.String(grpcHostFlag),
		port: c.Int(grpcPortFlag),
		sp:   sp,
	}
}

func (s *GRPC) Stat(ctx context.Context, r *pb.StatRequest) (*pb.StatReply, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	if len(md.Get("download-id")) == 0 || md.Get("download-id")[0] == "" {
		return nil, errors.Errorf("No download id provided")
	}
	downloadID := md.Get("download-id")[0]
	st := s.sp.GetIfExists(downloadID)
	if st == nil {
		return &pb.StatReply{
			Downloaded: 0,
			Status:     pb.StatReply_NOT_STARTED,
			Rate:       0,
			Length:     0,
		}, nil
	} else {
		var status pb.StatReply_Status
		switch st.status {
		case Pending:
			status = pb.StatReply_PENDING
			break
		case Active:
			status = pb.StatReply_ACTIVE
			break
		case Done:
			status = pb.StatReply_DONE
			break
		case Failed:
			status = pb.StatReply_FAILED
			break
		}
		return &pb.StatReply{
			Downloaded: st.bytesWritten,
			Status:     status,
			Rate:       st.Rate(),
			Length:     st.length,
		}, nil
	}
}

func (s *GRPC) StatStream(r *pb.StatRequest, ss pb.DownloadProgress_StatStreamServer) error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for range ticker.C {
		if ss.Context().Err() != nil {
			return errors.Wrapf(ss.Context().Err(), "Got context error")
		}
		rep, err := s.Stat(ss.Context(), r)
		if err != nil {
			return errors.Wrapf(err, "Failed to get stat")
		}
		err = ss.Send(rep)
		if err != nil {
			return errors.Wrapf(err, "Failed to send stat")
		}
		if rep.GetStatus() == pb.StatReply_DONE || rep.GetStatus() == pb.StatReply_FAILED {
			return nil
		}
	}
	return nil
}

func (s *GRPC) Serve() error {
	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return errors.Wrap(err, "Failed to listen to tcp connection")
	}
	s.ln = ln
	var opts []grpc.ServerOption
	gs := grpc.NewServer(opts...)
	pb.RegisterDownloadProgressServer(gs, s)
	log.Infof("Serving GRPC at %v", addr)
	return gs.Serve(ln)
}

func (s *GRPC) Close() {
	log.Info("Closing GRPC")
	defer func() {
		log.Info("GRPC closed")
	}()
	if s.ln != nil {
		s.ln.Close()
	}
}
