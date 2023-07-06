package main

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/tdex-network/tdex-daemon/cmd/migration/service"
	v0migration "github.com/tdex-network/tdex-daemon/cmd/migration/v0-v1"
)

var (
	versionFlag          = "source-version"
	defaultSourceVersion = "v0"
	allowedVersions      = map[string]service.Service{
		"v0": v0migration.NewService(),
	}

	version = "dev"
	commit  = "none"
	date    = "unknown"

	app = &cobra.Command{
		Use:                "migration",
		Short:              "migration service",
		Long:               "this service translates a tdex-daemon datadir and storage from a version to the next major one",
		Version:            formatVersion(),
		RunE:               action,
		DisableFlagParsing: true,
		SilenceUsage:       true,
		SilenceErrors:      true,
	}

	sourceVersion string
)

func init() {
	app.Flags().StringVarP(&sourceVersion, versionFlag, "", defaultSourceVersion, "the version of the daemon to be migrated to next major one")
}

func main() {
	if err := app.Execute(); err != nil {
		log.Fatal(err)
	}
}

func action(cmd *cobra.Command, args []string) (err error) {
	if len(args) > 0 && args[0] == "--version" {
		fmt.Println(formatVersion())
		return
	}

	migrationSvc, ok := allowedVersions[sourceVersion]
	if !ok {
		return fmt.Errorf("migration from version %s not supported", sourceVersion)
	}

	start := time.Now()
	log.Info("starting migration...")

	defer func(start time.Time) {
		if err == nil {
			elapsedTime := time.Since(start).Seconds()
			log.Infof("migration ended in %fs", elapsedTime)
		}
	}(start)

	if err = migrationSvc.Migrate(); err != nil {
		return
	}

	return
}

func formatVersion() string {
	return fmt.Sprintf(
		"Version: %s\nCommit: %s\nDate: %s",
		version, commit, date,
	)
}
