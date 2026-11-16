package main

import (
	"flag"
	"log"

	"termini/internal/app"
	"termini/internal/platform/explorer"
)

// main 启动 Termini 桌面应用。
func main() {
	workingDirectory := flag.String("working-directory", "", "终端启动目录") // 右键菜单指定的终端启动目录，空值使用用户主目录。
	administrator := flag.Bool("administrator", false, "启动 (管理员)") // 管理员菜单先提权，再启动终端。
	flag.Parse()
	if *administrator {
		if err := explorer.RunAsAdministrator(*workingDirectory); err != nil {
			log.Fatal(err)
		}
		return
	}
	if err := app.Run(*workingDirectory); err != nil {
		log.Fatal(err)
	}
}
