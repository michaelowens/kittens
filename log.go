package main

import (
	"github.com/mgutz/ansi"
	"log"
	"os"
)

var (
	d     =  ansi.ColorCode("white+b:magenta")
	i     =  ansi.ColorCode("white+b:blue")
	w     =  ansi.ColorCode("white+b:red")
	v     =  ansi.ColorCode("white+b:black")
	r     =  ansi.ColorCode("reset")
)

var (
	INFO  =  log.New(os.Stdout, d+"[kittens]"+r+" "+i+"INFO"+r+" ", 0)
	WARN  =  log.New(os.Stdout, d+"[kittens]"+r+" "+w+"WARN"+r+" ", 0)
	VERB  =  log.New(os.Stdout, d+"[kittens]"+r+" "+v+"VERB"+r+" ", 0)
)

// Print info statements
func info(i interface{}) {
	INFO.Print(i)
}

// Printf info statements
func infof(s string, i interface{}) {
	INFO.Printf(s, i)
}

// Print warning statements
func warn(i interface{}) {
	WARN.Print(i)
}

// Printf warning statements
func warnf(s string, i interface{}) {
	WARN.Printf(s, i)
}

// Print verbose statements if debug is true
func verb(i interface{}) {
	if (config.Debug) {
		VERB.Print(i)
	}
}

// Printf verbose statements if debug is true
func verbf(s string, i interface{}) {
	if (config.Debug) {
		VERB.Printf(s, i)
	}
}