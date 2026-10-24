// Команда vyshkat запускает программу
//
//	go run ./cmd/vyshkat         интерактивное меню
//	go run ./cmd/vyshkat -demo   демо-сценарий со всеми случаями
package main

import (
	"flag"
	"fmt"
	"os"

	"vyshkat/internal/console"
	"vyshkat/internal/di"
)

func main() {
	demo := flag.Bool("demo", false, "прогнать демо-сценарий и выйти")
	flag.Parse()

	container, err := di.NewContainer(di.DefaultConfig(os.Stdin, os.Stdout))
	if err == nil {
		err = container.Invoke(func(app *console.App) error {
			if *demo {
				app.RunDemo()
				return nil
			}
			return app.Run()
		})
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "ошибка:", err)
		os.Exit(1)
	}
}
