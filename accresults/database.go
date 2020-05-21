package accresults

import (
    "fmt"
    "io/ioutil"
)

type Database struct {
    Sessions map[string]*Session
}

func LoadDatabase(resultsPath string) (*Database, error) {
    var db = &Database{ make(map[string]*Session) }
    
    files, err := ioutil.ReadDir(resultsPath)
    if err != nil {
        return nil, err
    }
    
    for _, f := range files {
        filename := f.Name()
        db.Sessions[filename], err = LoadSessionFromFile(resultsPath + filename)
    }
    
    return db, err
}
