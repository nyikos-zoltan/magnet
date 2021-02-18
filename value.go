package magnet

func (m *Magnet) Value(v interface{}) {
	node := NewValueNode(v, m)
	m.providerMap[node.provides] = node
}
