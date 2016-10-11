package main

import (
    "flag"
    "log"
    "time"
    "github.com/kardianos/service"
    "github.com/cyberporthos/memory_access"
)

var logger service.Logger

// Program structures.
//  Define Start and Stop methods.
type program struct {
    exit chan struct{}
}

func (p *program) Start(s service.Service) error {
    if service.Interactive() {
        logger.Info("Running in terminal.")
    } else {
        memory_access.SetNoFeedback()
    }
    p.exit = make(chan struct{})

    // Start should not block. Do the actual work async.
    go p.run()
    return nil
}
func (p *program) run() error {
    ticker_duration, ticker_change_duration_chan := memory_access.GetTimerSeconds()
    // logger.Infof("I'm running %v with interval of %d sec.", service.Platform(), ticker_duration)
    ticker := time.NewTicker(time.Duration(ticker_duration) * time.Second)
    for {
        select {
        case <-ticker.C:
            memory_access.Run()
        case <-ticker_change_duration_chan:
            ticker.Stop()
            return p.run()
        case <-p.exit:
            ticker.Stop()
            return nil
        }
    }
}
func (p *program) Stop(s service.Service) error {
    // Any work in Stop should be quick, usually a few seconds at most.
    if service.Interactive() {
      logger.Info("I'm Stopping!")
    }
    close(p.exit)
    return nil
}

// Service setup.
//   Define service config.
//   Create the service.
//   Setup the logger.
//   Handle service controls (optional).
//   Run the service.
func main() {
    svcFlag := flag.String("service", "", "Control the system service.")
    flag.Parse()

    svcConfig := &service.Config{
        Name:        "MemoryIntegration",
        DisplayName: "Memory integration",
        Description: "Software used for Memory integration",
    }

    prg := &program{}
    s, err := service.New(prg, svcConfig)
    if err != nil {
        log.Fatal(err)
    }
    errs := make(chan error, 5)
    logger, err = s.Logger(errs)
    if err != nil {
        log.Fatal(err)
    }

    go func() {
        for {
            err := <-errs
            if err != nil {
                log.Print(err)
            }
        }
    }()

    if len(*svcFlag) != 0 {
        err := service.Control(s, *svcFlag)
        if err != nil {
            log.Printf("Valid actions: %q\n", service.ControlAction)
            log.Fatal(err)
        }
        return
    }
    err = s.Run()
    if err != nil {
        logger.Error(err)
    }
}
