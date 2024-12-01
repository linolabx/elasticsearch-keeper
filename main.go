package main

import (
	"log"
	"os"

	"github.com/linolabx/cli_helpers/helpers"
	"github.com/linolabx/cli_helpers/plugins/_es"
	"github.com/linolabx/cli_helpers/plugins/_gin"
	"github.com/linolabx/cli_helpers/plugins/_redis"
	"github.com/linolabx/cli_helpers/plugins/_zerolog"
	"github.com/linolabx/lino_redis"

	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"
)

func main() {
	godotenv.Load()
	app := &cli.App{}

	app.Name = "Elastisearch Keeper"
	app.Usage = "Elasticsearch keeper service, for receiving updates to the configuration file"
	app.Version = "alpha"
	app.Authors = []*cli.Author{
		{Name: "GeekTR", Email: "geektr@sxxfuture.net"},
	}

	app.Writer = os.Stderr
	app.Commands = []*cli.Command{
		GetServeCommand(),
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func GetServeCommand() *cli.Command {
	zerologPs := _zerolog.NewZeroLogPS()
	ginPs := _gin.NewGinPS()
	redisPs := _redis.NewRedisPS().EnablePrefixFlag()
	esPs := _es.NewEsPS()

	return helpers.CommandHelper{
		Name:  "serve",
		Usage: "start the keeper service",
		Flags: []cli.Flag{
			(&helpers.FlagHelper{
				Name:  "synonyms-config-dir",
				Usage: "elasticsearch synonyms config directory",
				Value: "/usr/share/elasticsearch/config/synonyms",
			}).StringFlag(),
			(&helpers.FlagHelper{
				Name:     "api-key",
				Usage:    "api key",
				Required: true,
			}).StringFlag(),
			&cli.BoolFlag{
				Name:  "dev",
				Usage: "dev mode",
				Value: false,
			},
		},
		Plugins: []helpers.CommandPlugin{zerologPs, ginPs, redisPs, esPs},
		Action: func(c *cli.Context) error {
			logger := zerologPs.GetInstance().With().Logger()

			sCtx := &ServerContext{
				Logger:        &logger,
				GinAddr:       ginPs.GetValue(),
				Redis:         lino_redis.NewLinoRedis(redisPs.GetInstance(), redisPs.GetPrefix()).Fork("es-keeper"),
				Es:            esPs.GetClient(),
				EsIndexPrefix: esPs.GetIndexPrefix(),

				SynonymsConfigDir: c.String("synonyms-config-dir"),
				ApiKey:            c.String("api-key"),

				DevMode: c.Bool("dev"),
			}

			RunServer(sCtx)

			return nil
		},
	}.Export()
}
