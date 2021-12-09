package clingy

type paramsTracker struct {
	count int
	list  []*param
}

func (pt *paramsTracker) getCount() int { return len(pt.list) }

func (pt *paramsTracker) hasErrors() bool {
	for _, p := range pt.list {
		if p != nil && p.err != nil {
			return true
		}
	}
	return false
}

func (pt *paramsTracker) params(cb func(*param)) {
	for _, p := range pt.list {
		cb(p)
	}
}

func (pt *paramsTracker) Break() {
	pt.list = append(pt.list, nil)
}

func (pt *paramsTracker) include(p *param) {
	pt.list = append(pt.list, p)
	pt.count++
}
