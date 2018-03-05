package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/jroimartin/gocui"
	messages "github.com/whereswaldon/arbor/messages"
)

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

var currentView int = 0

func up(g *gocui.Gui, v *gocui.View) error {
	currentView++
	_, err := g.SetCurrentView(fmt.Sprintf("%d", currentView))
	return err
}

func down(g *gocui.Gui, v *gocui.View) error {
	currentView--
	_, err := g.SetCurrentView(fmt.Sprintf("%d", currentView))
	return err
}

func main() {
	if len(os.Args) < 2 {
		log.Println("Usage: " + os.Args[0] + " <host:port>")
		return
	}
	ui, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Println("Unable to launch ui", err)
		return
	}
	defer ui.Close()

	queries := make(chan string)
	layoutManager := NewList(messages.NewStore(), queries)
	ui.Highlight = true
	ui.Cursor = true
	ui.SelFgColor = gocui.ColorGreen
	ui.SetManager(layoutManager)

	conn, err := net.Dial("tcp", os.Args[1])
	if err != nil {
		log.Println("Unable to connect", err)
		return
	}
	go HandleConn(conn, layoutManager, ui)
	go HandleRequests(conn, queries)

	if err := ui.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	if err := ui.SetKeybinding("", gocui.KeyArrowUp, gocui.ModNone, up); err != nil {
		log.Panicln(err)
	}

	if err := ui.SetKeybinding("", gocui.KeyArrowDown, gocui.ModNone, down); err != nil {
		log.Panicln(err)
	}

	if err = ui.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Println("error with ui:", err)
	}
}
