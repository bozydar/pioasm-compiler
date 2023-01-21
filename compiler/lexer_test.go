package compiler

import (
	"testing"
)

func Test_item_String(t *testing.T) {
	t.Run("Returns proper error", func(t *testing.T) {
		val := "This is error"
		i := lexItem{
			typ: itemError,
			val: val,
		}

		want := "This is error"
		if got := i.String(); got != want {
			t.Errorf("`%v` != `%v`", got, want)
		}
	})

	t.Run("Returns proper EOF", func(t *testing.T) {
		i := lexItem{
			typ: itemEOF,
			val: "",
		}

		want := "EOF"
		if got := i.String(); got != want {
			t.Errorf("`%v` != `%v`", got, want)
		}
	})

	t.Run("Returns proper label", func(t *testing.T) {
		i := lexItem{
			typ: itemLabel,
			val: ".label1",
		}

		want := "\".label1\""
		if got := i.String(); got != want {
			t.Errorf("`%v` != `%v`", got, want)
		}
	})

	t.Run("Ignores comments", func(t *testing.T) {
		input := `; comment
// comment
;comment
//comment`
		_, itemsCh := lex("test", input)
		items := make([]lexItem, 0)
		for {
			item, ok := <-itemsCh
			if !ok {
				break
			}
			t.Logf("%#v\n", item)
			items = append(items, item)
		}
		if l := len(items); l > 1 {
			t.Errorf("Too many items: %d", l)
		}
		if items[0].typ != itemEOF {
			t.Fail()
		}
	})

	t.Run("Compresses EOLs", func(t *testing.T) {
		input := `


; comment


`
		lexer, itemsCh := lex("test", input)
		items := make([]lexItem, 0)

		for {
			item, ok := <-itemsCh
			if !ok {
				break
			}
			items = append(items, item)
			t.Logf("%#v\n", item)
		}
		t.Logf("%#v", lexer)

		if items[0].typ != itemEOF {
			t.Error()
		}
	})

	t.Run("Emits instruction", func(t *testing.T) {
		input := `JMP
wait
in
out
push
pull
mov
irq
set`
		_, itemsCh := lex("test", input)
		items := make([]lexItem, 0)
		for {
			item, ok := <-itemsCh
			if !ok {
				break
			}
			t.Logf("%#v\n", item)
			items = append(items, item)
		}
		if items[0].typ != itemInstrJMP {
			t.Error()
		}
		if items[2].typ != itemInstrWAIT {
			t.Error()
		}
		if items[4].typ != itemInstrIN {
			t.Error()
		}
		if items[6].typ != itemInstrOUT {
			t.Error()
		}
		if items[8].typ != itemInstrPUSH {
			t.Error()
		}
		if items[10].typ != itemInstrPULL {
			t.Error()
		}
		if items[12].typ != itemInstrMOV {
			t.Error()
		}
		if items[14].typ != itemInstrIRQ {
			t.Error()
		}
		if items[16].typ != itemInstrSET {
			t.Error()
		}
	})

	t.Run("Directive with empty comment.", func(t *testing.T) {
		input := `
.define A 1 ;
.define B 2;;;;;;
`

		_, itemsCh := lex("test", input)
		items := make([]lexItem, 0)
		for {
			item, ok := <-itemsCh
			if !ok {
				break
			}
			t.Logf("%#v\n", item)
			items = append(items, item)
		}
		if items[0].typ != itemDirDefine {
			t.Error()
		}
		if items[1].typ != itemSymbol {
			t.Error()
		}
		if items[2].typ != itemNumber {
			t.Error()
		}
		if items[3].typ != itemEOL {
			t.Error()
		}
		if items[4].typ != itemDirDefine {
			t.Error()
		}
		if items[5].typ != itemSymbol {
			t.Error()
		}
		if items[6].typ != itemNumber {
			t.Error()
		}
		if items[7].typ != itemEOL {
			t.Error()
		}

	})

	t.Run("Parse 10-based integer.", func(t *testing.T) {
		input := `
23
`

		_, itemsCh := lex("test", input)
		items := make([]lexItem, 0)
		for {
			item, ok := <-itemsCh
			if !ok {
				break
			}
			t.Logf("%#v\n", item)
			items = append(items, item)
		}
		if items[0].typ != itemNumber {
			t.Error()
		}
	})

	//	t.Run("Parse 16-based integer.", func(t *testing.T) {
	//		input := `
	//0xffee
	//`
	//
	//		_, itemsCh := lex("test", input)
	//		items := make([]lexItem, 0)
	//		for {
	//			item, ok := <-itemsCh
	//			if !ok {
	//				break
	//			}
	//			t.Logf("%#v\n", item)
	//			items = append(items, item)
	//		}
	//		if items[0].typ != itemNumber {
	//			t.Error()
	//		}
	//	})

	t.Run("Emits identifier", func(t *testing.T) {
		input := `
.program ws2812
.side_set 1

.define public T1 2
.define public T2 5
.define public T3 3

.lang_opt python sideset_init = pico.PIO.OUT_HIGH
.lang_opt python out_init = pico.PIO.OUT_HIGH
.lang_opt python out_shiftdir = 1

.wrap_target
bitloop:
	out x, 1 side 0 [T3 - 1] ; Side-set still takes place when instruction stalls
	jmp !x do_zero side 1 [T1 - 1] ; Branch on the bit we shifted out. Positive pulse
	do_one:
	jmp bitloop side 1 [T2 - 1] ; Continue driving high, for a long pulse
	do_zero:
	nop side 0 [T2 - 1] ; Or drive low, for a short pulse
.wrap
`
		lexer, items := lex("test", input)
		for {
			item, ok := <-items
			if !ok {
				break
			}
			t.Logf("%#v\n", item)
		}
		t.Logf("%#v", lexer)
	})

}
