package student_memento

import (
	"github.com/owlint/goddd"
)

type StudentMemento struct {
	goddd.EventStream

	grade string
}

func NewStudentMemento(grade string) *StudentMemento {
	s := StudentMemento{}
	s.SetGrade(grade)

	return &s
}

func (s *StudentMemento) ObjectID() string {
	return "ObjectID"
}

func (s *StudentMemento) SetGrade(grade string) {
	s.AddEvent(s, "GradeSet", GradeSet{grade})
}

func (s *StudentMemento) OnGradeSet(event GradeSet) error {
	s.grade = event.Grade
	return nil
}

func (s *StudentMemento) ReloadMemento(memento Memento) error {
	s.grade = memento.Grade
	return nil
}
