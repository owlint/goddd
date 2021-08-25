package student

//go:generate msgp
//go:generate goddd_gen Student

type GradeSet struct {
    Grade string
}
