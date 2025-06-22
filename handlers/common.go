package handlers

import (
	"log"
	"os"
)

var logger = log.New(os.Stdout, "", log.LstdFlags)
var errLogger = log.New(os.Stderr, "", log.LstdFlags)
