package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	csi "github.com/container-storage-interface/spec/lib/go/csi/v0"
	"github.com/opensds/nbp/csi/server/plugin"
	"github.com/opensds/nbp/csi/server/plugin/opensds"
	"github.com/opensds/nbp/csi/util"
	"github.com/spf13/cobra"

	_ "github.com/opensds/nbp/driver/iscsi"
	_ "github.com/opensds/nbp/driver/rbd"
)

var (
	csiEndpoint         string
	opensdsEndpoint     string
	opensdsAuthStrategy string
)

func init() {
	flag.Set("logtostderr", "true")
}

func main() {

	flag.CommandLine.Parse([]string{})

	cmd := &cobra.Command{
		Use:   "OpenSDS",
		Short: "CSI based OpenSDS driver",
		Run: func(cmd *cobra.Command, args []string) {
			handle()
		},
	}

	cmd.Flags().AddGoFlagSet(flag.CommandLine)

	cmd.PersistentFlags().StringVar(&csiEndpoint, "csiEndpoint", "", "CSI Endpoint")
	cmd.PersistentFlags().StringVar(&opensdsEndpoint, "opensdsEndpoint", "", "OpenSDS Endpoint")
	cmd.PersistentFlags().StringVar(&opensdsAuthStrategy, "opensdsAuthStrategy", "", "OpenSDS Auth Strategy")

	cmd.ParseFlags(os.Args[1:])
	if err := cmd.Execute(); err != nil {
		log.Fatalf("failed to execute: %v", err)
		os.Exit(1)
	}

	os.Exit(0)
}

func handle() {

	// Set Env
	os.Setenv("CSI_ENDPOINT", csiEndpoint)
	os.Setenv("OPENSDS_ENDPOINT", opensdsEndpoint)
	os.Setenv("OPENSDS_AUTH_STRATEGY", opensdsAuthStrategy)

	// Get CSI Endpoint Listener
	lis, err := util.GetCSIEndPointListener()
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// New Grpc Server
	s := grpc.NewServer()

	// Register CSI Service
	var defaultplugin plugin.Service = &opensds.Plugin{}
	conServer := &server{plugin: defaultplugin}
	csi.RegisterIdentityServer(s, conServer)
	csi.RegisterControllerServer(s, conServer)
	csi.RegisterNodeServer(s, conServer)

	// Register reflection Service
	reflection.Register(s)

	// Remove sock file
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs)
	go func() {
		for sig := range sigs {
			if sig == syscall.SIGKILL ||
				sig == syscall.SIGQUIT ||
				sig == syscall.SIGHUP ||
				sig == syscall.SIGTERM ||
				sig == syscall.SIGINT {
				log.Println("exit to serve")
				if lis.Addr().Network() == "unix" {
					sockfile := lis.Addr().String()
					os.RemoveAll(sockfile)
					log.Printf("remove sock file: %s", sockfile)
				}
				os.Exit(0)
			}
		}
	}()

	// Serve Plugin Server
	log.Printf("start to serve: %s", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
