package xlib

import (
	"fmt"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
	"go.uber.org/zap"

	"yuki.no/muon/rect"
)

var (
	Conn             *xgb.Conn
	DefaultScreen    *xproto.ScreenInfo
	DefaultColormap  xproto.Colormap
	RootWindow       xproto.Window
	DeleteWindowAtom xproto.Atom
	NameAtom         xproto.Atom
	ProtocolsAtom    xproto.Atom
	TransientForAtom xproto.Atom
)

func init() {
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)

	var err error
	if Conn, err = xgb.NewConn(); err != nil {
		zap.S().Fatal(err)
	}

	DefaultScreen = xproto.Setup(Conn).DefaultScreen(Conn)
	RootWindow = DefaultScreen.Root
	DefaultColormap = DefaultScreen.DefaultColormap

	DeleteWindowAtom = MustInternAtom("WM_DELETE_WINDOW")
	NameAtom = MustInternAtom("WM_NAME")
	ProtocolsAtom = MustInternAtom("WM_PROTOCOLS")
	TransientForAtom = MustInternAtom("WM_TRANSIENT_FOR")
}

func MapWindow(id xproto.Window) {
	xproto.MapWindow(Conn, id)
}

func SetFocus(id xproto.Window) {
	xproto.SetInputFocus(Conn, xproto.InputFocusPointerRoot, id, xproto.TimeCurrentTime)
}

func Raise(id xproto.Window) {
	xproto.ConfigureWindow(Conn, id, xproto.ConfigWindowStackMode, []uint32{uint32(xproto.StackModeAbove)})
}

func SetBorderColor(id xproto.Window, color uint32) {
	xproto.ChangeWindowAttributes(Conn, id, xproto.CwBorderPixel, []uint32{color})
}

func SetBorderWidth(id xproto.Window, width int) {
	xproto.ConfigureWindow(Conn, id, xproto.ConfigWindowBorderWidth, []uint32{uint32(width)})
}

func SetGeometry(id xproto.Window, geometry rect.Rect) {
	var mask = xproto.ConfigWindowX | xproto.ConfigWindowY | xproto.ConfigWindowHeight | xproto.ConfigWindowWidth
	xproto.ConfigureWindow(Conn, id, uint16(mask), []uint32{
		uint32(geometry.X), uint32(geometry.Y), uint32(geometry.W), uint32(geometry.H),
	})
}

func ConfigureFromGeometry(id xproto.Window, borderWidth int, geometry rect.Rect) {
	configureNotify := xproto.ConfigureNotifyEvent{
		Event:            id,
		Window:           id,
		AboveSibling:     0,
		X:                int16(geometry.X),
		Y:                int16(geometry.Y),
		Width:            uint16(geometry.W),
		Height:           uint16(geometry.H),
		BorderWidth:      uint16(borderWidth),
		OverrideRedirect: false,
	}

	xproto.SendEvent(Conn, false, id, xproto.EventMaskStructureNotify, string(configureNotify.Bytes()))
}

func ConfigureFromRequest(event xproto.ConfigureRequestEvent) {
	var values []uint32
	var mask uint16

	if event.ValueMask&xproto.ConfigWindowX != 0 {
		mask |= xproto.ConfigWindowX
		values = append(values, uint32(event.X))
	}

	if event.ValueMask&xproto.ConfigWindowY != 0 {
		mask |= xproto.ConfigWindowY
		values = append(values, uint32(event.Y))
	}

	if event.ValueMask&xproto.ConfigWindowWidth != 0 {
		mask |= xproto.ConfigWindowWidth
		values = append(values, uint32(event.Width))
	}

	if event.ValueMask&xproto.ConfigWindowHeight != 0 {
		mask |= xproto.ConfigWindowHeight
		values = append(values, uint32(event.Height))
	}

	if event.ValueMask&xproto.ConfigWindowBorderWidth != 0 {
		mask |= xproto.ConfigWindowBorderWidth
		values = append(values, uint32(event.BorderWidth))
	}

	if event.ValueMask&xproto.ConfigWindowSibling != 0 {
		mask |= xproto.ConfigWindowSibling
		values = append(values, uint32(event.Sibling))
	}

	if event.ValueMask&xproto.ConfigWindowStackMode != 0 {
		mask |= xproto.ConfigWindowStackMode
		values = append(values, uint32(event.StackMode))
	}

	xproto.ConfigureWindow(Conn, event.Window, mask, values)
}

func HexColor(color string) uint32 {
	var r, g, b uint16
	fmt.Sscanf(color, "#%02x%02x%02x", &r, &g, &b)
	r *= 0x101
	g *= 0x101
	b *= 0x101

	reply, err := xproto.AllocColor(Conn, DefaultColormap, r, g, b).Reply()
	if err != nil {
		return 0
	}

	return reply.Pixel
}

func MustInternAtom(name string) xproto.Atom {
	atom, err := InternAtom(name)
	if err != nil {
		zap.S().Fatal(err)
	}

	return atom
}

func InternAtom(name string) (xproto.Atom, error) {
	reply, err := xproto.InternAtom(Conn, true, uint16(len(name)), name).Reply()

	return reply.Atom, err
}

func GetProperty(window xproto.Window, atom xproto.Atom) (*xproto.GetPropertyReply, error) {
	return xproto.GetProperty(Conn, false, window, atom, xproto.GetPropertyTypeAny, 0, (1<<32)-1).Reply()
}

func GetPropertyValue(window xproto.Window, atom xproto.Atom) ([]byte, error) {
	reply, err := GetProperty(window, atom)

	return reply.Value, err
}

func GetAtoms(window xproto.Window, atom xproto.Atom) []xproto.Atom {
	if reply, err := GetProperty(window, atom); err == nil {
		atoms := make([]xproto.Atom, reply.ValueLen)
		for value, i := reply.Value, 0; len(value) >= 4; i++ {
			atoms[i] = xproto.Atom(xgb.Get32(value))
			value = value[4:]
		}

		return atoms
	}

	return nil
}

func GetString(window xproto.Window, atom xproto.Atom) string {
	if property, err := GetPropertyValue(window, atom); err == nil {
		return string(property)
	}

	return ""
}

func GetWindow(window xproto.Window, atom xproto.Atom) xproto.Window {
	if property, err := GetPropertyValue(window, atom); err == nil && len(property) != 0 {
		return xproto.Window(xgb.Get32(property))
	}

	return 0
}

func ClientMessage(window xproto.Window, args ...uint32) {
	ClientMessage32(window, ProtocolsAtom, args...)
}

func ClientMessage8(window xproto.Window, atom xproto.Atom, args ...uint8) {
	var buffer = make([]uint8, 20)
	for i := 0; i < 20; i++ {
		if i >= len(args) {
			break
		}

		buffer[i] = args[i]
	}

	message := xproto.ClientMessageEvent{
		Window:   window,
		Sequence: 0,
		Format:   8,
		Type:     atom,
		Data:     xproto.ClientMessageDataUnionData8New(buffer),
	}

	xproto.SendEvent(Conn, false, window, xproto.EventMaskNoEvent, string(message.Bytes()))
}

func ClientMessage16(window xproto.Window, atom xproto.Atom, args ...uint16) {
	var buffer = make([]uint16, 10)
	for i := 0; i < 10; i++ {
		if i >= len(args) {
			break
		}

		buffer[i] = args[i]
	}

	message := xproto.ClientMessageEvent{
		Window:   window,
		Sequence: 0,
		Format:   16,
		Type:     atom,
		Data:     xproto.ClientMessageDataUnionData16New(buffer),
	}

	xproto.SendEvent(Conn, false, window, xproto.EventMaskNoEvent, string(message.Bytes()))
}

func ClientMessage32(window xproto.Window, atom xproto.Atom, args ...uint32) {
	var buffer = make([]uint32, 5)
	for i := 0; i < 5; i++ {
		if i >= len(args) {
			break
		}

		buffer[i] = args[i]
	}

	message := xproto.ClientMessageEvent{
		Window:   window,
		Sequence: 0,
		Format:   32,
		Type:     atom,
		Data:     xproto.ClientMessageDataUnionData32New(buffer),
	}

	xproto.SendEvent(Conn, false, window, xproto.EventMaskNoEvent, string(message.Bytes()))
}
