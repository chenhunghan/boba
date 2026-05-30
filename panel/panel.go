// Package panel is a coordinate-routing module for multi-panel TUI
// layouts. It answers one question: given a screen `(x, y)` and a
// tree of nested splits, which leaf Panel contains that point and
// what are the panel-content-local coordinates?
//
// The package is intentionally narrow. It does not render anything,
// does not know about box.Box or any body-stack package, and does
// not know about focus. A leaf Panel carries an ID (caller-defined
// type, typically a focus enum) so callers can dispatch to whatever
// per-panel logic they own. Borders on each leaf are subtracted
// before returning local coordinates, so callers that wrap their
// panels in a bordered Box receive coordinates ready to feed into
// their button.Stack.HitTest / navcard.Stack.HitTest.
//
// The layout tree has two node kinds:
//
//   - Panel[F]: a leaf with a focus ID, a Size in its parent's axis,
//     and Borders subtracted from local coords.
//   - Split[F]: a container with an Axis (Horizontal or Vertical),
//     a list of child Nodes, and a Size in its own parent's axis.
//
// Sizes are interpreted in the parent's axis. A child with Size=0 is
// a "remainder" — it gets whatever size is left after fixed-Size
// siblings consume their share. At most one remainder is permitted
// per Split; more than one is a structural bug and HitTest panics.
//
// The root node's own Size is ignored (the root has no parent); it
// always fills the (width, height) passed to HitTest.
//
// Typical usage:
//
//	var layout = panel.Split[focusArea]{
//	    Axis: panel.Horizontal,
//	    Children: []panel.Node[focusArea]{
//	        panel.Panel[focusArea]{ID: focusSidebar,    Size: 5,  Borders: panel.Borders{Top: 1, Left: 1, Right: 1}},
//	        panel.Panel[focusArea]{ID: focusContent, Size: 30, Borders: panel.Borders{Top: 1, Left: 1, Right: 1}},
//	        panel.Panel[focusArea]{ID: focusMain,      Size: 0,  Borders: panel.Borders{Top: 1, Left: 1, Right: 1}},
//	    },
//	}
//	hit := panel.HitTest(layout, screenWidth, screenHeight, x, y)
//	if hit.Found {
//	    switch hit.Panel { /* dispatch using hit.LocalX, hit.LocalY */ }
//	}
package panel

// Axis is the direction a Split arranges its children.
type Axis int

const (
	// Horizontal lays children out left-to-right; child Size is width.
	Horizontal Axis = iota
	// Vertical lays children out top-to-bottom; child Size is height.
	Vertical
)

// Borders are the rows and columns to subtract from the panel's
// outer rect when computing content-local coordinates. Set Top=1 and
// Left=1 (etc.) for a panel whose body sits inside a one-cell box
// border so callers receive body-local coords directly.
type Borders struct {
	Top, Right, Bottom, Left int
}

// Node is one node in the layout tree. The interface is closed —
// only Panel[F] and Split[F] satisfy it. The unexported method
// prevents external types from implementing Node, which keeps
// HitTest's switch exhaustive.
type Node[F comparable] interface {
	nodeImpl()
}

// Panel is a leaf in the layout tree. ID is the caller-defined focus
// or panel identifier returned in Hit. Size is the panel's extent in
// its parent's axis (width for a Horizontal parent, height for a
// Vertical parent); Size=0 marks the panel as the parent's remainder.
// Borders are subtracted from local coords before the Hit is returned.
type Panel[F comparable] struct {
	ID      F
	Size    int
	Borders Borders
}

func (Panel[F]) nodeImpl() {}

// Split is a container that arranges Children along Axis. Size is
// the Split's own extent in its parent's axis; Size=0 marks the
// Split as the parent's remainder.
type Split[F comparable] struct {
	Axis     Axis
	Children []Node[F]
	Size     int
}

func (Split[F]) nodeImpl() {}

// Hit is the result of a coordinate query. Found is false when the
// query coordinate is outside the root rect or falls in an unfilled
// gap inside an empty Split. When Found is true, Panel is the leaf
// ID and (LocalX, LocalY) are the panel-content-local coordinates
// (Borders already subtracted) — ready for a per-panel HitTest.
type Hit[F comparable] struct {
	Found          bool
	Panel          F
	LocalX, LocalY int
}

// HitTest descends root and returns the leaf Panel containing
// (x, y) in screen coordinates. width and height are the screen
// rectangle the root fills (root's own Size is ignored). Returns
// Hit{} (Found=false) if (x, y) is outside that rectangle.
//
// Panics if any Split contains more than one Size=0 child — that's
// a structural error in the layout, surfaced loudly.
func HitTest[F comparable](root Node[F], width, height, x, y int) Hit[F] {
	if x < 0 || y < 0 || x >= width || y >= height {
		return Hit[F]{}
	}
	return hit[F](root, 0, 0, width, height, x, y)
}

// hit is the recursive worker. (ox, oy) is the origin of n's rect in
// screen coords; (w, h) is the size of n's rect. (x, y) is the screen
// coord being queried (already known to be inside (ox, oy, w, h) by
// the caller).
func hit[F comparable](n Node[F], ox, oy, w, h, x, y int) Hit[F] {
	switch v := n.(type) {
	case Panel[F]:
		return Hit[F]{
			Found:  true,
			Panel:  v.ID,
			LocalX: x - ox - v.Borders.Left,
			LocalY: y - oy - v.Borders.Top,
		}
	case Split[F]:
		return hitSplit(v, ox, oy, w, h, x, y)
	}
	panic("panel: unknown Node kind")
}

func hitSplit[F comparable](s Split[F], ox, oy, w, h, x, y int) Hit[F] {
	avail := w
	if s.Axis == Vertical {
		avail = h
	}
	sizes := distribute(s.Children, avail)
	cursor := 0
	for i, child := range s.Children {
		sz := sizes[i]
		switch s.Axis {
		case Horizontal:
			if x >= ox+cursor && x < ox+cursor+sz {
				return hit[F](child, ox+cursor, oy, sz, h, x, y)
			}
		case Vertical:
			if y >= oy+cursor && y < oy+cursor+sz {
				return hit[F](child, ox, oy+cursor, w, sz, x, y)
			}
		}
		cursor += sz
	}
	return Hit[F]{}
}

// distribute assigns a concrete size to each child given the
// available extent in the parent's axis. The remainder child (Size=0)
// receives `avail - sumOfFixedSizes`; all others receive their Size.
// Panics if more than one child has Size=0.
func distribute[F comparable](children []Node[F], avail int) []int {
	sizes := make([]int, len(children))
	fixed := 0
	remainderIdx := -1
	for i, c := range children {
		sz := nodeSize[F](c)
		if sz == 0 {
			if remainderIdx >= 0 {
				panic("panel: at most one Size=0 (remainder) child per Split")
			}
			remainderIdx = i
			continue
		}
		sizes[i] = sz
		fixed += sz
	}
	if remainderIdx >= 0 {
		sizes[remainderIdx] = avail - fixed
	}
	return sizes
}

// nodeSize returns the parent-axis size declared on a node.
func nodeSize[F comparable](n Node[F]) int {
	switch v := n.(type) {
	case Panel[F]:
		return v.Size
	case Split[F]:
		return v.Size
	}
	panic("panel: unknown Node kind")
}
