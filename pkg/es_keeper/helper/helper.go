package helper

import (
	"github.com/linolabx/cli_helpers/helpers"
	"github.com/linolabx/elasticsearch-keeper/pkg/es_keeper"
	"github.com/urfave/cli/v2"
)

type ESKeeperPS struct {
	ESKeeperDSN helpers.FlagHelper

	esKeeper *es_keeper.ESKeeper

	initialized bool
}

func (this *ESKeeperPS) SetPrefix(prefix string) *ESKeeperPS {
	this.ESKeeperDSN.Prefix = prefix
	return this
}

func (this *ESKeeperPS) SetCategory(category string) *ESKeeperPS {
	this.ESKeeperDSN.Category = category
	return this
}

func (this *ESKeeperPS) HandleCommand(cmd *cli.Command) error {
	cmd.Flags = append(cmd.Flags, this.ESKeeperDSN.StringFlag())
	return nil
}

func (this *ESKeeperPS) HandleContext(cCtx *cli.Context) error {
	ek, err := es_keeper.NewESKeeper(this.ESKeeperDSN.StringValue(cCtx))
	if err != nil {
		return err
	}
	this.esKeeper = ek
	this.initialized = true
	return nil
}

func (this *ESKeeperPS) GetInstance() *es_keeper.ESKeeper {
	if !this.initialized {
		panic("ESKeeperPS is not initialized")
	}
	return this.esKeeper
}

func NewESKeeperPS() *ESKeeperPS {
	return &ESKeeperPS{
		ESKeeperDSN: helpers.FlagHelper{
			Name:     "es-keeper-dsn",
			Required: true,
			Category: "datasource",
			Usage:    "Elasticsearch Keeper DSN, e.g. http://localhost:9200?api_key=1234567890",
		},
	}
}
