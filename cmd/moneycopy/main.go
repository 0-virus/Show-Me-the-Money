package main

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"math/rand"
	"moneycopy/internal/terminal"
	"os"
	"time"
)

const version = "1.0.3"

func run() error {
	screen, err := tcell.NewScreen()
	if err != nil {
		return err
	}
	if err = screen.Init(); err != nil {
		return err
	}
	defer screen.Fini()
	screen.SetTitle("가만히 있어도 돈이 복사가 되면 얼마나 좋을까")
	screen.EnableMouse(tcell.MouseButtonEvents)
	screen.EnablePaste()
	app := terminal.New(screen, time.Now, rand.New(rand.NewSource(time.Now().UnixNano())))
	app.Run()
	return nil
}
func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Println("Money Copy", version)
		return
	}
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "터미널을 시작하지 못했습니다:", err)
		fmt.Fprintln(os.Stderr, "Windows Terminal 또는 macOS Terminal에서 실행해 주세요.")
		os.Exit(1)
	}
}
