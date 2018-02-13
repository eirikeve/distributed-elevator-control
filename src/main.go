package main

import (
	"os"
)

func main() {
	for {
		err := run()
		if err != nil {
			elevlog.log(err)
		} else {
			os.Exit(0)
		}

	}
}

func run() error {
	err := something()
	if err != nil {
		return err
	}
	// etc
}
