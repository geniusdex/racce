package main

import (
    "log"
    "github.com/geniusdex/racce/accresults"
)

func main() {
    log.Printf("Populating database...")
    session, err := accresults.LoadDatabase("results/")
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Starting server...")
    log.Fatal(RunServer(session))
}
