package main

import (
	"log"
	"net"
	"os"

	"github.com/jroimartin/gocui"
	"github.com/whereswaldon/arbor/client/clientio"
	"github.com/whereswaldon/arbor/messages"
)

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
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

	layoutManager, queries := NewList(NewTree(messages.NewStore()))
	msgs := make(chan *messages.Message)
	ui.Highlight = true
	ui.Cursor = true
	ui.SelFgColor = gocui.ColorGreen
	ui.SetManager(layoutManager)

	conn, err := net.Dial("tcp", os.Args[1])
	if err != nil {
		log.Println("Unable to connect", err)
		return
	}
	go clientio.HandleNewMessages(conn, msgs)
	go func() {
		for newMsg := range msgs {
			layoutManager.Add(newMsg)
			layoutManager.UpdateMessage(newMsg.UUID)
			//redraw
			ui.Update(func(*gocui.Gui) error { return nil })
		}
	}()
	go clientio.HandleRequests(conn, queries)

	if err := ui.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	if err := ui.SetKeybinding("", gocui.KeyArrowUp, gocui.ModNone, layoutManager.Up); err != nil {
		log.Panicln(err)
	}

	if err := ui.SetKeybinding("", gocui.KeyArrowDown, gocui.ModNone, layoutManager.Down); err != nil {
		log.Panicln(err)
	}

	if err = ui.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Println("error with ui:", err)
	}
}
