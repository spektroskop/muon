package muon

import (
	"fmt"

	"github.com/BurntSushi/xgb/xproto"
	"go.uber.org/zap"

	"yuki.no/muon/rect"
	"yuki.no/muon/xlib"
)

type Window struct {
	Id        xproto.Window
	Name      string
	Geometry  rect.Rect
	Protocols map[xproto.Atom]bool
}

func (win *Window) String() string {
	if win.Name != "" {
		return fmt.Sprintf("%s—0x%08x", win.Name, win.Id)
	}

	return fmt.Sprintf("0x%08x", win.Id)
}

func NewWindow(mr *Manager, id xproto.Window) *Window {
	win := new(Window)

	win.Id = id
	win.Name = xlib.GetString(id, xlib.NameAtom)
	win.Protocols = make(map[xproto.Atom]bool)
	for _, atom := range xlib.GetAtoms(id, xlib.ProtocolsAtom) {
		win.Protocols[atom] = true
	}

	zap.S().Infow("window", "id", win)

	return win
}
