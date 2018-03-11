package main

import (
	"log"
	"net"
	"os"

	"github.com/jroimartin/gocui"
	"github.com/pkg/profile"
	"github.com/whereswaldon/arbor/client/clientio"
	"github.com/whereswaldon/arbor/messages"
)

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func main() {
	defer profile.Start().Stop()
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

	layoutManager, queries, outbound := NewList(NewTree(messages.NewStore()))
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
			layoutManager.UpdateLeaf(newMsg.UUID)
			//redraw
			ui.Update(func(*gocui.Gui) error { return nil })
		}
	}()
	go clientio.HandleRequests(conn, queries)
	go clientio.HandleOutbound(conn, outbound)

	type keybinding struct {
		viewId  string
		key     interface{} // must be a rune or gocui.Key
		mod     gocui.Modifier
		handler func(*gocui.Gui, *gocui.View) error
	}
	bindings := []keybinding{
		{"", gocui.KeyCtrlC, gocui.ModNone, quit},
		{"", gocui.KeyArrowUp, gocui.ModNone, layoutManager.CursorUp},
		{"", gocui.KeyArrowDown, gocui.ModNone, layoutManager.CursorDown},
		{"", gocui.KeyArrowLeft, gocui.ModNone, layoutManager.CursorLeft},
		{"", gocui.KeyArrowRight, gocui.ModNone, layoutManager.CursorRight},
		/*
			{"", 'q', gocui.ModNone, quit},
			{"", 'k', gocui.ModNone, layoutManager.CursorUp},
			{"", 'j', gocui.ModNone, layoutManager.CursorDown},
			{"", 'h', gocui.ModNone, layoutManager.CursorLeft},
			{"", 'l', gocui.ModNone, layoutManager.CursorRight},
		*/
		{"", gocui.KeyEnter, gocui.ModNone, layoutManager.BeginReply},
	}

	for _, binding := range bindings {
		log.Println("registering ", binding.key)
		if err := ui.SetKeybinding(binding.viewId, binding.key, binding.mod, binding.handler); err != nil {
			log.Panicln(err)
		}
	}

	if err = ui.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Println("error with ui:", err)
	}
}
