package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"time"

	"github.com/docker/docker/client"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

type serveCfg struct {
	rpcAddr        string
	traefikGnoFile string
	hostPWD        string

	masterBackupFile string
	snapshotsDir     string
}

func (c *serveCfg) RegisterFlags(fs *flag.FlagSet) {
	if os.Getenv("HOST_PWD") == "" {
		os.Setenv("HOST_PWD", os.Getenv("PWD"))
	}

	if os.Getenv("SNAPSHOTS_DIR") == "" {
		os.Setenv("SNAPSHOTS_DIR", "./backups/snapshots")
	}

	if os.Getenv("RPC_URL") == "" {
		os.Setenv("RPC_URL", "http://rpc.portal.gno.local:26657")
	}

	if os.Getenv("PROM_ADDR") == "" {
		os.Setenv("PROM_ADDR", ":9090")
	}

	if os.Getenv("TRAEFIK_GNO_FILE") == "" {
		os.Setenv("TRAEFIK_GNO_FILE", "./traefik/gno.yml")
	}

	if os.Getenv("MASTER_BACKUP_FILE") == "" {
		os.Setenv("MASTER_BACKUP_FILE", "./backups/backup.jsonl")
	}

	fs.StringVar(&c.rpcAddr, "rpc", os.Getenv("RPC_URL"), "tendermint rpc url")
	fs.StringVar(&c.traefikGnoFile, "traefik-gno-file", os.Getenv("TRAEFIK_GNO_FILE"), "traefik gno file")
	fs.StringVar(&c.hostPWD, "pwd", os.Getenv("HOST_PWD"), "host pwd (for docker usage)")
	fs.StringVar(&c.masterBackupFile, "master-backup-file", os.Getenv("MASTER_BACKUP_FILE"), "master txs backup file path")
	fs.StringVar(&c.snapshotsDir, "snapshots-dir", os.Getenv("SNAPSHOTS_DIR"), "snapshots directory")
}

func newServeCmd(_ commands.IO) *commands.Command {
	cfg := &serveCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "serve",
			ShortUsage: "serve [flags]",
		},
		cfg,
		func(ctx context.Context, args []string) error {
			return execServe(ctx, cfg, args)
		},
	)
}

func execServe(ctx context.Context, cfg *serveCfg, args []string) error {
	dockerClient, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	portalLoop := &snapshotter{}

	// Serve monitoring
	go func() {
		s := &monitoringService{
			portalLoop: portalLoop,
		}

		for portalLoop.url == "" {
			time.Sleep(time.Second * 1)
		}

		go s.recordMetrics()

		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(os.Getenv("PROM_ADDR"), nil) //nolint:all
	}()

	// the loop
	for {
		portalLoop, err = NewSnapshotter(dockerClient, config{
			snapshotsDir:     cfg.snapshotsDir,
			masterBackupFile: cfg.masterBackupFile,
			rpcAddr:          cfg.rpcAddr,
			hostPWD:          cfg.hostPWD,
			traefikGnoFile:   cfg.traefikGnoFile,
		})
		if err != nil {
			return err
		}

		err = StartPortalLoop(ctx, portalLoop, false)
		if err != nil {
			logrus.WithError(err).Error()
		}
		time.Sleep(time.Second * 120)
	}
}
