package handler

import (
	"github.com/muhfajar/riuh/api"
	"log"
)

func main() {
	err := api.Routes().Run()
	if err != nil {
		log.Fatal(err)
	}
}
