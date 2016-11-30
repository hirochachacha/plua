package codegen

type scope struct {
	symbols map[string]link // linkLocal or linkUpval
	labels  map[string]label

	outer   *scope
	savedSP int

	doClose bool // generate CLOSE(JMP) op when closeScope called

	nlocals int // if r >= nlocals then r is tmp variable

	endPC int
}

func (s *scope) root() *scope {
	scope := s
	for {
		if scope.outer == nil {
			return scope
		}

		scope = scope.outer
	}
}

func (s *scope) resolveLabel(name string) (label, bool) {
	scope := s
	for {
		l, ok := scope.labels[name]
		if ok {
			return l, true
		}

		scope = scope.outer
		if scope == nil {
			return label{}, false
		}
	}
}

func (s *scope) declare(name string, l link) {
	s.symbols[name] = l
}

func (s *scope) resolveLocal(name string) (link, bool) {
	scope := s
	for {
		l, ok := scope.symbols[name]
		if ok {
			return l, true
		}

		scope = scope.outer
		if scope == nil {
			return link{}, false
		}
	}
}
