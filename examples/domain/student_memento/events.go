package student_memento

//go:generate msgp
//go:generate goddd_gen StudentMemento

type Memento struct {
	ID    string
	Grade string
}

type GradeSet struct {
	Grade string
}
