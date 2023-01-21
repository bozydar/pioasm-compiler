package compiler

import (
	"testing"
)

func Test_Compile(t *testing.T) {
	t.Run("Returns defines on the file level", func(t *testing.T) {
		source := `.define A 1
.define public BBB 2 ; comment`
		ast, _ := Compile(source, &Options{})

		sourceOut := ast.ToSource()

		if sourceOut != `.define A 1
.define BBB 2
` {
			t.Logf(sourceOut)
			t.Errorf("Regenerated source is different")
		}
	})

	t.Run("Error if syntax error. Case 1", func(t *testing.T) {
		source := `bullshit`

		ast, e := Compile(source, &Options{})

		if ast != nil {
			t.Errorf("%#v", ast)
		}
		if e == nil || e.offset != 1 || e.line != 1 {
			t.Errorf("%#v", e)
		}
	})

	t.Run("Error if syntax error. Case 2", func(t *testing.T) {
		source := `

 bullshit
`

		ast, e := Compile(source, &Options{})

		if ast != nil {
			t.Errorf("%#v", ast)
		}
		if e == nil || e.offset != 2 || e.line != 3 {
			t.Errorf("%#v", e)
		}
	})

	t.Run("Throws an error 2", func(t *testing.T) {
		source := `
     
 bullshit
`

		ast, e := Compile(source, &Options{})

		if ast != nil {
			t.Errorf("%#v", ast)
		}
		if e == nil || e.offset != 2 || e.line != 3 {
			t.Errorf("%#v", e)
		}
	})

	t.Run("Error if redeclare .define. Case 1.", func(t *testing.T) {
		source := `
.define A 1
.define A 2
`
		ast, e := Compile(source, &Options{})

		if ast != nil {
			t.Errorf("%#v", ast)
		}
		if e == nil || e.offset != 1 || e.line != 3 {
			t.Errorf("%#v", e)
		}
	})

	t.Run("Error if redeclare .define. Case 2.", func(t *testing.T) {
		source := `
.define A 1
.define B 2
.define A 3
`
		ast, e := Compile(source, &Options{})

		if ast != nil {
			t.Errorf("%#v", ast)
		}
		if e == nil || e.offset != 1 || e.line != 4 {
			t.Errorf("%#v", e)
		}
	})

	t.Run("Error if redeclare .define in a program. Case 1.", func(t *testing.T) {
		source := `
.define A 1

.program test
.define B 4
.define A 3
`
		ast, e := Compile(source, &Options{})

		if ast != nil {
			t.Errorf("%#v", ast)
		}
		if e == nil || e.offset != 1 || e.line != 6 {
			t.Errorf("%#v", e)
		}
	})

	t.Run("Error if redeclare .define in a program. Case 2.", func(t *testing.T) {
		source := `
.define A 1

.program testA
.define B 4

.program testB
.define A 3
`
		ast, e := Compile(source, &Options{})

		if ast != nil {
			t.Errorf("%#v", ast)
		}
		if e == nil || e.offset != 1 || e.line != 8 {
			t.Errorf("%#v", e)
		}
	})

	t.Run("No error if the same .define in separate programs. Case 1.", func(t *testing.T) {
		source := `
.program testA
.define B 4

.program testB
.define B 3
`
		ast, e := Compile(source, &Options{})

		if ast == nil {
			t.Errorf("%#v", ast)
		}
		if e != nil {
			t.Errorf("%#v", e)
		}
	})

	t.Run("Error if a second program with the same id. Case 1.", func(t *testing.T) {
		source := `
.program testA
.program testA
`
		ast, e := Compile(source, &Options{})

		if ast != nil {
			t.Errorf("%#v", ast)
		}
		if e == nil || e.offset != 1 || e.line != 3 {
			t.Errorf("%#v", e)
		}
	})

	t.Run("Error if a second program with the same id. Case 2.", func(t *testing.T) {
		source := `
.program testA
.program testB
.program testA
`
		ast, e := Compile(source, &Options{})

		if ast != nil {
			t.Errorf("%#v", ast)
		}
		if e == nil || e.offset != 1 || e.line != 4 {
			t.Errorf("%#v", e)
		}
	})

	t.Run("Returns defines on the file level", func(t *testing.T) {
		source := `
.define A 1

.program test

.define B 1
.define public BBB 2 ; comment`
		ast, _ := Compile(source, &Options{})

		sourceOut := ast.ToSource()

		if sourceOut != `.define A 1
.program test
.define B 1
.define BBB 2
` {
			t.Logf(sourceOut)
			t.Errorf("Regenerated source is different")
		}
	})

	t.Run("Defines with comments. Case 1.", func(t *testing.T) {
		source := `
.define A 60 ;
.define B A + 13; comment
`

		ast, e := Compile(source, &Options{})

		sourceOut := ast.ToSource()

		if e != nil {
			t.Errorf("%#v", e)
		}

		if sourceOut != `.define A 60
.define B A + 13 ; = 73
` {
			t.Logf(sourceOut)
			t.Errorf("Regenerated source is different")
		}
	})

}
