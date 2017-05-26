package main

import (
	"fmt"
	"os"

	"github.com/Ragnis/go-translate"
	"github.com/Ragnis/go-translate/example/resid"
)

func main() {
	d := translate.DefaultDomain
	d.SetVersionHash(resid.VersionHash)
	d.MustLoadStrings("strings/en.pak.json")
	d.MustLoadStrings("strings/et.pak.json")

	lang, err := d.Language("en")
	if err != nil {
		panic(err)
	}

	name := ""
	if len(os.Args) > 1 {
		name = os.Args[1]
	}

	if name == "" {
		fmt.Printf(lang.String(resid.Greeting))
	} else {
		fmt.Printf(lang.String(resid.GreetingWithName), name)
	}

	fmt.Println("")
}
