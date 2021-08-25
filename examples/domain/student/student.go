package student

import (
    "github.com/owlint/goddd"
)

type Student struct {
    goddd.EventStream

    grade string
}

func NewStudent(grade string) *Student {
    s := Student{}
    s.SetGrade(grade)


    return &s
}

func (s *Student) ObjectID() string {
    return "ObjectID"
}

func (s *Student) SetGrade(grade string) {
    s.AddEvent(s, "GradeSet", GradeSet{grade})
}

func (s *Student) OnGradeSet(event GradeSet) error {
    s.grade = event.Grade
    return nil
}

