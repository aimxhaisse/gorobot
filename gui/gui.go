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
	Frame		*gtk.GtkFrame
	VBox		*gtk.GtkVBox
	Text		*gtk.GtkTextView
	UserInput	*GoRobotGUIInput
}

type GoRobotGUIInput struct {
	Input		*gtk.GtkTextView
	Button		*gtk.GtkButton
	Box		*gtk.GtkHBox
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

func NewGoRobotGUIInput() (*GoRobotGUIInput) {
	input := GoRobotGUIInput{
		Input: gtk.TextView(),
		Button: gtk.ButtonWithLabel("Send"),
		Box: gtk.HBox(false, 1),
	}

	input.Box.Add(input.Input)
	input.Box.Add(input.Button)
	return &input
}

func (gui *GoRobotGUI) GetTab(name string) (*GoRobotGUITab) {
	tab, ok := gui.Tabs[name]
	if !ok {
		tab = new(GoRobotGUITab)
		tab.Frame = gtk.Frame(name)
		tab.VBox = gtk.VBox(false, 1)
		tab.Text = gtk.TextView()
		tab.UserInput = NewGoRobotGUIInput()
		tab.Frame.Add(tab.VBox)
		tab.VBox.Add(tab.Text)
		tab.VBox.Add(tab.UserInput.Box)
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
