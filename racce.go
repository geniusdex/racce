package main

import (
    "log"
    "github.com/geniusdex/racce/accresults"
)

func main() {
    session, err := accresults.LoadDatabase("results/")
    if err != nil {
        log.Fatal(err)
    }
    log.Fatal(RunServer(session))
}
