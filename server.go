package main

import (
    "log"
    "net/http"
    "os"
    "time"
    "github.com/urfave/cli/v2"
    "github.com/op/go-logging"
)

const LOG_FORMAT = "%{color}%{time:2006/01/02 15:04:05 -07:00 MST} [%{level:.6s}] %{shortfile} : %{color:reset}%{message}"

func runServer(c *cli.Context) error {

    ///////////////////////////////// LOGGER
    // NewLogger creates instance of logger that should be used
    // in all server handlers and routines. The idea is to have
    // unified style of logging - logger is configured only once
    // and at one place
    backend := logging.NewLogBackend(os.Stderr, "", 0)
    format := logging.MustStringFormatter(LOG_FORMAT)
    backendFormatter := logging.NewBackendFormatter(backend, format)

    backendLeveled := logging.AddModuleLevel(backendFormatter)
    logLevel, err := logging.LogLevel(c.String("log-level"))
    if err != nil {
        log.Fatalf("Cannot create logger for level %s (%v)", c.String("log-level"), err)
        return err
        //os.Exit(1)
    }
    backendLeveled.SetLevel(logLevel, "")

    logging.SetBackend(backendLeveled)
    logger := logging.MustGetLogger("server")

    logger.Infof("Starting BSBigMap server")

    ////////////////////////////////// TILE PROVIDERS

    logger.Infof("Reading providers")
    providers, err := readProviders()
    if err != nil {
        logger.Errorf("Providers config error: %s", err)
        return err
    }
    for provider := range providers {
        logger.Infof("- %s %s", provider, providers[provider].Name)
    }

    ////////////////////////////////// QUEUE
    queue, err := NewQueue(
            logger,
            c.String("queue-dir"),
            c.Duration("queue-validity"),
            c.Duration("queue-monitor-interval"),
        )
    if err != nil {
        return err
    }

    ////////////////////////////////// HTTP HANDLERS
    http.Handle("/stitcher", &HandlerParams{logger, providers, &HandlerStitcher{logger, providers, queue}})

    http.Handle("/map", &HandlerParams{logger, providers, &HandlerMap{logger, providers}})

    http.Handle("/queue", &HandlerQueue{logger, queue})

    // server static content
    //fs := http.FileServer(http.Dir("."))
    //http.Handle("/static/", http.StripPrefix("/static/", fs))
    fs := http.FileServer(http.Dir(queue.dir))
    http.Handle("/queue/", http.StripPrefix("/" + queue.dir + "/", fs))


    http.Handle("/", &HandlerRoot{logger, providers})


    ////////////////////////////////// SERVER

    logger.Infof("Listening on %s...", c.String("bind-address"))
    err = http.ListenAndServe(c.String("bind-address"), nil)
    return err
}

func FatalOnError(err error, msg string, args ...interface{}) {
    if err != nil {
        log.Fatalf(msg, args...)
        os.Exit(1)
    }
}

func main() {
    app := cli.NewApp()

    app.Name = "BSBigMap Server"
    app.Version = "1.0"
    app.Authors = []*cli.Author{
        {
            Name:  "Michal Nezerka",
            Email: "michal.nezerka@gmail.com",
        },
    }
    app.Usage = "Stitch map tiles into single PNG image"
    app.Action = runServer
    app.Flags = []cli.Flag{
        &cli.StringFlag{
            Name:   "bind-address",
            Aliases: []string{"b"},
            Usage:  "Listen address for API HTTP endpoint",
            Value:  "0.0.0.0:9090",
            EnvVars: []string{"BIND_ADDRESS"},
        },
        &cli.StringFlag{
            Name:   "log-level",
            Aliases: []string{"l"},
            Usage:  "Logging level",
            Value:  "INFO",
            EnvVars: []string{"LOG_LEVEL"},
        },
        &cli.PathFlag{
            Name:   "queue-dir",
            Aliases: []string{"q"},
            Usage:  "Directory used for storing queued files",
            Value:  "queue",
            EnvVars: []string{"QUEUE_DIR"},
        },
        &cli.DurationFlag{
            Name: "queue-validity",
            Usage: "The maximal time for request (generated image) to be kept in queue",
            Value: time.Hour * 24 * 2,
            EnvVars: []string{"QUEUE_VALIDITY"},
        },
        &cli.DurationFlag{
            Name: "queue-monitor-interval",
            Usage: "The interval in which queue checks new requests to be processed",
            Value: time.Second * 5,
            EnvVars: []string{"QUEUE_MONITOR_INTERVAL"},
        },

    }

    err := app.Run(os.Args)
    if err != nil {
        FatalOnError(err, "Failed to start server: %s", err)
    }
}
