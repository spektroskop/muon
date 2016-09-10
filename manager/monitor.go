package manager

import (
	"fmt"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/Sirupsen/logrus"
	"github.com/spektroskop/muon/node"
	"github.com/spektroskop/muon/util"
)

type Arranger func(*Monitor, int) []util.Geometry

type Monitor struct {
	util.Geometry

	Name string

	Windows *node.Node
	Focus   *node.Node
	Layout  *node.Node

	Ratio    float64
	Border   int
	Gap      int
	Roots    int
	Mirrored bool
}

func (m *Monitor) Reset() {
	m.Ratio = 0.65
	m.Border = 5
	m.Gap = 1
	m.Roots = 1
	m.Mirrored = false
}

func NewMonitor(name string, geometry util.Geometry) *Monitor {
	m := &Monitor{
		Name:     name,
		Geometry: geometry,
		Windows:  node.New(nil),
		Focus:    nil,
	}

	m.Reset()
	m.AddLayout(Standing)
	m.AddLayout(Laying)

	return m
}

func (m Monitor) String() string {
	return fmt.Sprintf("%s: %s", m.Name, m.Geometry.String())
}

func (m *Monitor) SetFocus(node *node.Node) {
	if m.Focus != nil {
		focus := m.Focus.Value.(*Window)
		focus.SetBorderColor(InactiveBorder)
	}

	window := node.Value.(*Window)
	window.SetBorderColor(ActiveBorder)
	window.Focus()

	m.Focus = node
}

func (m *Monitor) WindowFromId(id xproto.Window) *node.Node {
	for _, node := range m.Windows.Nodes() {
		if node.Value.(*Window).Id == id {
			return node
		}
	}

	return nil
}

func (m *Monitor) AddLayout(arranger Arranger) {
	if node := node.New(arranger); m.Layout == nil {
		m.Layout = node
	} else {
		m.Layout.Append(node)
	}
}

func (m *Monitor) Arrange() {
	logrus.Debugf("Arrange monitor `%s'", m.Name)

	arranger := m.Layout.Value.(Arranger)
	geom := arranger(m, m.Windows.Len())

	for i, node := range m.Windows.Nodes() {
		window := node.Value.(*Window)
		g := geom[i]

		logrus.Debugf(" 0x%08x %s", window.Id, g)
		window.SetGeometry(g)
		window.SetBorderWidth(m.Border)
	}
}
