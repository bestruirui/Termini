package main

import (
	"flag"
	"log"

	"termini/internal/app"
)

// main 启动 Termini 桌面应用。
func main() {
	workingDirectory := flag.String("working-directory", "", "终端启动目录") // 右键菜单指定的终端启动目录，空值使用用户主目录。
	flag.Parse()
	if err := app.Run(*workingDirectory); err != nil {
		log.Fatal(err)
	}
}
