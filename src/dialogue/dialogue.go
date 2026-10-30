package dialogue

func (c Catalog) Get(id string) (*Dialogue, bool) {
	d, ok := c[id]
	return d, ok
}
