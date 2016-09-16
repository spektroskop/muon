package manager

import (
	"fmt"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/Sirupsen/logrus"
	"github.com/spektroskop/muon/nd"
	"github.com/spektroskop/muon/util"
)

type Arranger func(*Monitor, int) []util.Geometry

type Monitor struct {
	util.Geometry

	Name string

	Root  *nd.Node
	Focus *nd.Node

	Layout *nd.Node

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
	m := &Monitor{Name: name, Geometry: geometry}

	m.Reset()
	m.AddLayout(Standing)
	m.AddLayout(Laying)

	return m
}

func (m Monitor) String() string {
	return fmt.Sprintf("%s: %s", m.Name, m.Geometry.String())
}

func (m *Monitor) SetFocus(node *nd.Node) {
	if m.Focus != nil {
		focus := m.Focus.Value.(*Window)
		focus.SetBorderColor(InactiveBorder)
	}

	WithWindowNode(node, func(window *Window) {
		window.SetBorderColor(ActiveBorder)
		window.Focus()
	})

	m.Focus = node
}

func (m *Monitor) NodeFromId(id xproto.Window) *nd.Node {
	for _, node := range m.Root.All() {
		if node.Value.(*Window).Id == id {
			return node
		}
	}

	return nil
}

func (m *Monitor) AddLayout(arranger Arranger) {
	if node := nd.New(arranger); m.Layout == nil {
		m.Layout = node
	} else {
		node.Link(m.Layout)
	}
}

func (m *Monitor) Arrange() {
	logrus.Debugf("Arrange monitor `%s'", m.Name)

	arranger := m.Layout.Value.(Arranger)
	geom := arranger(m, m.Root.Len())

	for i, node := range m.Root.All() {
		window := node.Value.(*Window)
		g := geom[i]

		logrus.Debugf(" 0x%08x %s", window.Id, g)
		window.SetGeometry(g)
		window.SetBorderWidth(m.Border)
	}
}
