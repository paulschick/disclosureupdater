package commands

import (
	"fmt"
	"github.com/paulschick/disclosureupdater/common/constants"
	"github.com/paulschick/disclosureupdater/common/logger"
	"github.com/paulschick/disclosureupdater/config"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"net/http"
	_ "net/http/pprof"
)

// Commander is the base struct for all commands run through viper CLI
// This stores the logger and profile.
// it also stores folder and file path information and provides startup and cleanup methods.
type Commander struct {
	log         *zap.Logger
	profile     string
	runProfiler bool
	commonDirs  *config.CommonDirs
}

// init Initializes the commander
// This can be run with a set of cli flags that will configure the environment.
// The program uses a default folder and configuration file location to store any modified values.
// --profile is used to specify the profile to use for the configuration.
// If no profile is specified, the default profile is used.
// --images is used to specify the folder to store the images.
// --disclosures is used to specify the folder to store the disclosures.
// --ocr is used to specify the folder to store the ocr files.
// --csv is used to specify the folder to store the csv files.
// --s3 is used to specify the folder to store the s3 files.
// --profile will run the profiler on port 6060
func (c *Commander) init(cc *cli.Context) {
	logger.InitLogger()
	c.log = logger.Logger
	err := c.initConfig(cc)
	if err != nil {
		c.log.Error("Error initializing configuration", zap.Error(err))
		return
	}
}

// initConfig initializes the configuration for the commander
// see config package for more information on the configuration setup
func (c *Commander) initConfig(cc *cli.Context) error {
	c.profile = cc.String("profile")
	c.runProfiler = cc.Bool("profile")

	if c.profile == "" {
		c.profile = constants.DefaultProfile
	}

	b, err := config.NewConfigurationBuilder(c.profile, cc)
	if err != nil {
		return err
	}

	err = b.Build()
	if err != nil {
		return err
	}

	c.commonDirs = b.GetCommonDirs()

	return nil
}

// cleanup is run after the command has completed
// This will sync the logger and perform any other cleanup tasks.
// If an error occurred during the command, it will be logged.
func (c *Commander) cleanup() {
	_ = c.log.Sync()
}

// RunProfiler runs the profiler on port 6060
// This will be used as a flag with any command.
func (c *Commander) RunProfiler() {
	go func() {
		fmt.Println(http.ListenAndServe("localhost:6060", nil))
	}()
}
