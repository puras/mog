package main

/**
* @project kuko
* @author <a href="mailto:he@puras.cn">Puras.He</a>
* @date 2021-08-16 21:06
 */

import (
	"fmt"

	"github.com/spf13/pflag"
	"mooko.net/mog/handler"
	"mooko.net/mog/pkg/config"
	"mooko.net/mog/pkg/db"
	"mooko.net/mog/pkg/server"
)

var (
	cfg = pflag.StringP("config", "c", "", "Product config file path.")
)

func reportPanic() {
	p := recover()
	if p == nil {
		return
	}
	err, ok := p.(error)
	if ok {
		fmt.Println("启动出错", err)
	}
}

func main() {
	defer reportPanic()
	pflag.Parse()

	err := config.Init(*cfg)
	if err != nil {
		panic(err)
	}

	// Init db
	db.InitDB()
	defer db.CloseDB()

	srv := server.New()
	srv.Start(handler.RegistryRoute)
}
