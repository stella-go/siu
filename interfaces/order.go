package interfaces

type Order interface {
	Order() int
}

type OrderSlice[e Order] []e

func (p OrderSlice[e]) Len() int {
	return len(p)
}

func (p OrderSlice[e]) Less(i, j int) bool {
	return p[i].Order() < p[j].Order()
}

func (p OrderSlice[e]) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
