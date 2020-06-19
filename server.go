package main

import (
    "log"
    "net/http"
    "os"
    "github.com/urfave/cli"
    "github.com/op/go-logging"
)

const LOG_FORMAT = "%{color}%{time:2006/01/02 15:04:05 -07:00 MST} [%{level:.6s}] %{shortfile} : %{color:reset}%{message}"

func runServer(c *cli.Context) {

    ///////////////// LOGGER instance
    // NewLogger creates instance of logger that should be used
    // in all server handlers and routines. The idea is to have
    // unified style of logging - logger is configured only once
    // and at one place
    backend := logging.NewLogBackend(os.Stderr, "", 0)
    format := logging.MustStringFormatter(LOG_FORMAT)
    backendFormatter := logging.NewBackendFormatter(backend, format)

    backendLeveled := logging.AddModuleLevel(backendFormatter)
    logLevel, err := logging.LogLevel(c.GlobalString("log-level"))
    if err != nil {
        log.Fatalf("Cannot create logger for level %s (%v)", c.GlobalString("log-level"), err)
        os.Exit(1)
    }
    backendLeveled.SetLevel(logLevel, "")

    logging.SetBackend(backendLeveled)
    logger := logging.MustGetLogger("server")

    logger.Infof("Starting BSBigMap server")

    ////////////////////////////////// PROVIDERS

    logger.Infof("Reading providers")
    providers, err := readProviders()
    if err != nil {
        logger.Errorf("Providers config error: %s", err)
    }
    for provider := range providers {
        logger.Infof("- %s %v", provider, providers[provider])
    }

    http.Handle("/", &HandlerRoot{logger, providers})

    logger.Infof("Listening on %s...", c.GlobalString("bind-address"))
    err = http.ListenAndServe(c.GlobalString("bind-address"), nil)
    FatalOnError(err, "Failed to bind on %s: ", c.GlobalString("bind-address"))
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
    app.Authors = []cli.Author{
        {
            Name:  "Michal Nezerka",
            Email: "michal.nezerka@gmail.com",
        },
    }
    app.Usage = "Stitch map tiles into single PNG image"
    app.Action = runServer
    app.Flags = []cli.Flag{
        cli.StringFlag{
            Name:   "bind-address,b",
            Usage:  "Listen address for API HTTP endpoint",
            Value:  "0.0.0.0:9090",
            EnvVar: "BIND_ADDRESS",
        },
        cli.StringFlag{
            Name:   "log-level,l",
            Usage:  "Logging level",
            Value:  "INFO",
            EnvVar: "LOG_LEVEL",
        },
    }

    app.Run(os.Args)
}
