package translator

type Registers struct {
	data []string
	dx   int
}

func NewRegisters() *Registers {
	return &Registers{
		make([]string, 16),
		0,
	}
}

func (R *Registers) claim(T *Translator) string {
	R.dx--
	if R.dx < 0 {
		reg := T.nextIdentifier()
		R.dx++
		return reg
	}
	return R.data[R.dx]
}

func (R *Registers) free(s string) {
	R.data[R.dx] = s
	R.dx++
	if R.dx >= len(R.data) {
		old := R.data
		R.data = make([]string, len(old)*2)
		copy(R.data, old)
	}
}
