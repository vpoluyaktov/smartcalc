package eval

import "fmt"

func EvalExpr(expr string, refResolver func(n int) (float64, error)) (float64, error) {
	toks, err := Lex(expr)
	if err != nil {
		return 0, err
	}
	p := &parser{toks: toks, refs: refResolver}
	v, err := p.parseExpr(0)
	if err != nil {
		return 0, err
	}
	if p.cur().Kind != tokEOF {
		return 0, fmt.Errorf("unexpected token: %s", p.cur().Text)
	}
	return v.v, nil
}
