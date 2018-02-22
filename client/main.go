package main

import (
	"fmt"
	ui "github.com/gizak/termui"
	"io"
	"net"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: " + os.Args[0] + " <host:port>")
		return
	}
	conn, err := net.Dial("tcp", os.Args[1])
	if err != nil {
		fmt.Println("Unable to connect", err)
		return
	}
	//	io.Copy(os.Stdout, conn)
	go handleConn(conn)

	err = ui.Init()
	if err != nil {
		fmt.Println("Unable to launch ui", err)
		return
	}
	defer ui.Close()
	parent := ui.NewPar("This is a parent message")
	parent.Height = 3

	current := ui.NewPar("This is this current message")
	current.Height = 4

	sibling := ui.NewPar("This is a sibling message")
	sibling.Height = 4

	child1 := ui.NewPar("This is a child message")
	child1.Height = 5

	child2 := ui.NewPar("This is another child message")
	child2.Height = 5

	input := ui.NewPar("This is an input")
	input.Height = 3

	vline := ui.NewPar("|")
	vline.Border = false
	vline.Height = 1

	vline2 := ui.NewPar("|")
	vline2.Border = false
	vline2.Height = 1

	vline3 := ui.NewPar("|")
	vline3.Border = false
	vline3.Height = 1

	vline4 := ui.NewPar("|")
	vline4.Border = false
	vline4.Height = 1

	ui.Body.AddRows(
		ui.NewRow(
			ui.NewCol(12, 0, parent),
		),
		ui.NewRow(
			ui.NewCol(6, 0, vline),
			ui.NewCol(5, 1, vline2),
		),
		ui.NewRow(
			ui.NewCol(6, 0, current),
			ui.NewCol(5, 1, sibling),
		),
		ui.NewRow(
			ui.NewCol(4, 0, vline3),
			ui.NewCol(4, 0, vline4),
		),
		ui.NewRow(
			ui.NewCol(4, 0, child1),
			ui.NewCol(4, 0, child2),
		),
		ui.NewRow(
			ui.NewCol(12, 0, input),
		),
	)

	ui.Body.Align()
	ui.Render(ui.Body)

	ui.Handle("/sys/kbd/q", func(ui.Event) {
		ui.StopLoop()
	})
	ui.Loop()
}

func handleConn(conn io.ReadWriteCloser) {

}
