package compiler

import "testing"

func Test_Compile_Expr(t *testing.T) {
	t.Run("Parses expression. Case 1.", func(t *testing.T) {
		source := `
.define A 1 + 2
`

		ast, e := Compile(source, &Options{})

		if e != nil {
			t.Errorf("%#v", e)
		}
		if ast.defines[0].name != "A" || ast.defines[0].expr == nil {
			t.Errorf("%#v", ast)
		}
	})

	t.Run("Parses expression. Case 2.", func(t *testing.T) {
		source := `
.define A 1 + 2 * 3
`

		ast, e := Compile(source, &Options{})

		sourceOut := ast.ToSource()

		if sourceOut != `.define A 1 + 2 * 3 ; = 7
` {
			t.Logf(sourceOut)
			t.Errorf("Regenerated source is different")
		}

		if e != nil {
			t.Errorf("%#v", e)
		}
	})

	t.Run("Parses expression with parens. Case 1.", func(t *testing.T) {
		source := `
.define A (1 + 2) * 3
`

		ast, e := Compile(source, &Options{})

		sourceOut := ast.ToSource()

		if sourceOut != `.define A (1 + 2) * 3 ; = 9
` {
			t.Logf(sourceOut)
			t.Errorf("Regenerated source is different")
		}

		if e != nil {
			t.Errorf("%#v", e)
		}
	})

	t.Run("Parses expression with parens. Case 2.", func(t *testing.T) {
		source := `
.define A (1 + 2) / (3 - 5) * (4 / 2)
`

		ast, e := Compile(source, &Options{})

		sourceOut := ast.ToSource()

		// NOTE Removing of not needed parens
		if sourceOut != `.define A (1 + 2) / (3 - 5) * 4 / 2 ; = -2
` {
			t.Logf(sourceOut)
			t.Errorf("Regenerated source is different")
		}

		if e != nil {
			t.Errorf("%#v", e)
		}
	})

	t.Run("Parses expression with parens. Case 3.", func(t *testing.T) {
		source := `
.define A (1 + 2) / (3 - 5) * (4 / 2)
`

		ast, e := Compile(source, &Options{})

		sourceOut := ast.ToSource()

		// NOTE Removing of not needed parens
		if sourceOut != `.define A (1 + 2) / (3 - 5) * 4 / 2 ; = -2
` {
			t.Logf(sourceOut)
			t.Errorf("Regenerated source is different")
		}

		if e != nil {
			t.Errorf("%#v", e)
		}
	})

	t.Run("Parses expression with defined symbol. Case 1.", func(t *testing.T) {
		source := `
.define T1 1
.define A T1 + 2
`

		ast, e := Compile(source, &Options{})

		sourceOut := ast.ToSource()

		if sourceOut != `.define T1 1
.define A T1 + 2 ; = 3
` {
			t.Logf(sourceOut)
			t.Errorf("Regenerated source is different")
		}

		if e != nil {
			t.Errorf("%#v", e)
		}
	})

	t.Run("Parses expression with defined symbol. Case 2.", func(t *testing.T) {
		source := `
.define T1 1

.program A
.define TA 2

.program B
.define TB1 T1 + 1
.define TB2 T1 + TB1 * 2
`

		ast, e := Compile(source, &Options{})

		sourceOut := ast.ToSource()

		if sourceOut != `.define T1 1
.program A
.define TA 2
.program B
.define TB1 T1 + 1 ; = 2
.define TB2 T1 + TB1 * 2 ; = 5
` {
			t.Logf(sourceOut)
			t.Errorf("Regenerated source is different")
		}

		if e != nil {
			t.Errorf("%#v", e)
		}
	})

	t.Run("Directive with comment. Case 1.", func(t *testing.T) {
		source := `
.define A 60
.define B 13 
.define C 0

.define A_AND_B A & B
.define A_OR_B A | B
.define A_XOR_B A ^ B
`

		ast, e := Compile(source, &Options{})

		sourceOut := ast.ToSource()

		if sourceOut != `.define A 60
.define B 13
.define C 0
.define A_AND_B A & B ; = 12
.define A_OR_B A | B ; = 61
.define A_XOR_B A ^ B ; = 49
` {
			t.Logf(sourceOut)
			t.Errorf("Regenerated source is different")
		}

		if e != nil {
			t.Errorf("%#v", e)
		}
	})

	t.Run("Parses expression with bit operators. Case 1.", func(t *testing.T) {
		source := `
.define A 60 ;
.define B 13 
.define C 0

.define A_AND_B A & B
.define A_OR_B A | B
.define A_XOR_B A ^ B
`

		ast, e := Compile(source, &Options{})

		sourceOut := ast.ToSource()

		if sourceOut != `.define A 60
.define B 13
.define C 0
.define A_AND_B A & B ; = 12
.define A_OR_B A | B ; = 61
.define A_XOR_B A ^ B ; = 49
` {
			t.Logf(sourceOut)
			t.Errorf("Regenerated source is different")
		}

		if e != nil {
			t.Errorf("%#v", e)
		}
	})
}
