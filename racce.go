package main

import (
    "log"
    "github.com/geniusdex/racce/accresults"
)

func main() {
    session, err := accresults.LoadSessionFromFile("200520_215813_R.json")
    if err != nil {
        log.Fatal(err)
    }
    log.Fatal(RunServer(session))
}
