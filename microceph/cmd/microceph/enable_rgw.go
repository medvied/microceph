package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/canonical/microcluster/microcluster"
	"github.com/spf13/cobra"

	"github.com/canonical/microceph/microceph/api/types"
	"github.com/canonical/microceph/microceph/ceph"
	"github.com/canonical/microceph/microceph/client"
)

type cmdEnableRGW struct {
	common             *CmdControl
	wait               bool
	flagPort           int
	flagSSLPort        int
	flagSSLCertificate string
	flagSSLPrivateKey  string
	flagTarget         string
}

func (c *cmdEnableRGW) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rgw [--port <port>] [--ssl-port <port>] [--ssl-certificate <certificate path>] [--ssl-private-key <private key path>] [--target <server>] [--wait <bool>]",
		Short: "Enable the RGW service on the --target server (default: this server)",
		RunE:  c.Run,
	}
	cmd.PersistentFlags().IntVar(&c.flagPort, "port", 80, "Service non-SSL port (default: 80)")
	cmd.PersistentFlags().IntVar(&c.flagSSLPort, "ssl-port", 443, "Service SSL port (default: 443)")
	cmd.PersistentFlags().StringVar(&c.flagSSLCertificate, "ssl-certificate", "", "Path to SSL certificate")
	cmd.PersistentFlags().StringVar(&c.flagSSLPrivateKey, "ssl-private-key", "", "Path to SSL private key")
	cmd.PersistentFlags().StringVar(&c.flagTarget, "target", "", "Server hostname (default: this server)")
	cmd.Flags().BoolVar(&c.wait, "wait", true, "Wait for rgw service to be up.")
	return cmd
}

// Run handles the enable rgw command.
func (c *cmdEnableRGW) Run(cmd *cobra.Command, args []string) error {
	m, err := microcluster.App(microcluster.Args{StateDir: c.common.FlagStateDir, Verbose: c.common.FlagLogVerbose, Debug: c.common.FlagLogDebug})
	if err != nil {
		return err
	}

	cli, err := m.LocalClient()
	if err != nil {
		return err
	}

	// sanity check: are ssl files in a place the microcephd can read?
	if c.flagSSLCertificate != "" {
		for _, sslFile := range []string{c.flagSSLCertificate, c.flagSSLPrivateKey} {
			if !strings.HasPrefix(sslFile, os.Getenv("SNAP_COMMON")) &&
				!strings.HasPrefix(sslFile, os.Getenv("SNAP_USER_COMMON")) {
				// print warning
				fmt.Println("Warning: SSL files might not be readable by daemon. It's recommended to use files in $SNAP_COMMON or $SNAP_USER_COMMON.")
			}
		}
	}

	jsp, err := json.Marshal(ceph.RgwServicePlacement{Port: c.flagPort, SSLPort: c.flagSSLPort, SSLCertificate: c.flagSSLCertificate, SSLPrivateKey: c.flagSSLPrivateKey})
	if err != nil {
		return err
	}

	req := &types.EnableService{
		Name:    "rgw",
		Wait:    c.wait,
		Payload: string(jsp[:]),
	}

	err = client.SendServicePlacementReq(context.Background(), cli, req, c.flagTarget)
	if err != nil {
		return err
	}

	return nil
}
