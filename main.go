package main

import (
	"context"
	"errors"
	"github.com/joho/godotenv"
	httpLib "github.com/kas2000/http"
	"github.com/kas2000/logger"
	"github.com/urfave/cli/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"time"
)

var (
	port   = ""
	dbUri  = ""
	dbName = ""
	env    = ""

	flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "config",
			Aliases:     []string{"c"},
			Usage:       "Load configuration from `FILE`",
			Required:    true,
			Destination: &env,
		},
	}
)

func parseEnv() error {
	err := godotenv.Overload(env)
	if err != nil {
		return err
	}
	port = os.Getenv("PORT")
	if port == "" {
		return errors.New("invalid port")
	}
	dbUri = os.Getenv("DB_URI")
	if dbUri == "" {
		return errors.New("invalid db uri")
	}
	dbName = os.Getenv("DB_NAME")
	if dbName == "" {
		return errors.New("invalid db name")
	}

	return nil
}

func main() {
	app := &cli.App{
		Name:      "Region Test Case",
		Usage:     "service-todo",
		UsageText: "go run main.go/service-todo --config FILE",
		Flags:     flags,
		Action:    run,
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}

func run(*cli.Context) error {
	log, _ := logger.New("debug")

	if err := parseEnv(); err != nil {
		log.Fatal("Error parsing .env file: " + err.Error())
	}

	mongoClient, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(dbUri))
	if err != nil {
		log.Fatal("couldn't connect to mongodb: " + err.Error())
	}
	defer func() {
		if err := mongoClient.Disconnect(context.TODO()); err != nil {
			log.Fatal(err.Error())
		}
	}()
	mongoDB := mongoClient.Database(dbName)
	collNames, _ := mongoDB.ListCollectionNames(context.TODO(), bson.M{})
	collectionsNames := make(map[string]int)
	for _, collName := range collNames {
		collectionsNames[collName]++
	}

	serverConfig := httpLib.Config{
		IsGatewayServer: false,
		PublicKey:       nil,
		Port:            port,
		ShutdownTimeout: time.Second * 20,
		GracefulTimeout: time.Second * 21,
		ApiVersion:      "v1",
		Timeout:         time.Second * 20,
		Logger:          log,
	}
	server := httpLib.NewServer(serverConfig)
	if err != nil {
		log.Fatal("couldn't instantiate server: " + err.Error())
	}

	server.ListenAndServe()
	return nil
}
