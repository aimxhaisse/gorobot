package main

import (
	"os"
	"gtk"
	"api"
	"flag"
	"log"
	"fmt"
)

type GoRobotGUI struct {
	Window		*gtk.GtkWindow
	Notebook	*gtk.GtkNotebook
	Tabs		map[string] *GoRobotGUITab
}

type GoRobotGUITab struct {
	Text		*gtk.GtkTextView
	Frame		*gtk.GtkFrame
}

func NewGoRobotGUI() *GoRobotGUI {
	gui := GoRobotGUI{
		Window: gtk.Window(gtk.GTK_WINDOW_TOPLEVEL),
		Notebook: gtk.Notebook(),
		Tabs: make(map[string] *GoRobotGUITab),
	}

	gui.Window.SetTitle("GoRobot GUI")
	gui.Window.Connect("destroy", func() {
		gtk.MainQuit()
	}, nil)

	gui.Window.Add(gui.Notebook)
	gui.Window.SetSizeRequest(400, 300)
	gui.Window.ShowAll()

	return &gui
}

func (gui *GoRobotGUI) GetTab(name string) (*GoRobotGUITab) {
	tab, ok := gui.Tabs[name]
	if !ok {
		tab = new(GoRobotGUITab)
		tab.Frame = gtk.Frame(name)
		tab.Text = gtk.TextView()
		tab.Text.SetEditable(false)
		tab.Text.SetCursorVisible(false)
		tab.Frame.Add(tab.Text)
		gui.Notebook.AppendPage(tab.Frame, gtk.Label(name))
		gui.Tabs[name] = tab
	}
	return tab
}

func (gui *GoRobotGUI) WriteToTab(tab *GoRobotGUITab, msg string) {
	var iter gtk.GtkTextIter
	buffer := tab.Text.GetBuffer()
	buffer.GetEndIter(&iter)
	buffer.Insert(&iter, msg)
}

func (gui *GoRobotGUI) HandleEvents() {
	_, chev := api.ImportFrom(*addr, "gui")

	for {
		e := <- chev
		if e.Type == api.E_PRIVMSG {
			if len(e.Channel) > 0 {
				tab := gui.GetTab(e.Channel)
				gui.WriteToTab(tab, fmt.Sprintf("%s: %s\n", e.User, e.Data))
			} else {
				tab := gui.GetTab(e.User)
				gui.WriteToTab(tab, fmt.Sprintf("%s: %s\n", e.User, e.Data))
			}
		}
		gui.Window.ShowAll()
	}

}

var addr = flag.String("s", "", "address:port of exported netchans")

func main() {
	flag.Parse()
	if *addr == "" {
		log.Exit("Usage : ./module -s addr:port")
	}
	gtk.Init(&os.Args)
	gui := NewGoRobotGUI()
	go gtk.Main()
	gui.HandleEvents()
}
