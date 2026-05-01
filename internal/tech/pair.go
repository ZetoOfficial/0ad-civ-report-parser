package tech

// ExpandPair returns the pair's top/bottom Technologies. ok=false
// (no error) if name is not a pair wrapper — caller falls through
// to plain-tech rendering.
func ExpandPair(c *Catalog, name string) (top, bottom *Technology, ok bool) {
	wrap, err := c.ByName(name)
	if err != nil {
		return nil, nil, false
	}
	if wrap.Top == "" || wrap.Bottom == "" {
		return nil, nil, false
	}
	t, err := c.ByName(wrap.Top)
	if err != nil {
		return nil, nil, false
	}
	b, err := c.ByName(wrap.Bottom)
	if err != nil {
		return nil, nil, false
	}
	return t, b, true
}
