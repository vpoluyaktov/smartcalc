package eval

import (
	"fmt"
	"math"
)

func (p *parser) cur() Token {
	if p.pos >= len(p.toks) {
		return Token{Kind: tokEOF}
	}
	return p.toks[p.pos]
}

func (p *parser) eat(k TokenKind) (Token, error) {
	t := p.cur()
	if t.Kind != k {
		return Token{}, fmt.Errorf("expected %v, got %s", k, t.Text)
	}
	p.pos++
	return t, nil
}

func (p *parser) parseExpr(minPrec int) (val, error) {
	left, err := p.parsePrefix()
	if err != nil {
		return val{}, err
	}

	for {
		t := p.cur()
		prec, rightAssoc := infixPrec(t.Kind)
		if prec < minPrec {
			break
		}
		p.pos++
		nextMin := prec + 1
		if rightAssoc {
			nextMin = prec
		}
		right, err := p.parseExpr(nextMin)
		if err != nil {
			return val{}, err
		}
		switch t.Kind {
		case tokPlus:
			if right.pct {
				left = val{v: left.v * (1 + right.v)}
			} else {
				left = val{v: left.v + right.v}
			}
		case tokMinus:
			if right.pct {
				left = val{v: left.v * (1 - right.v)}
			} else {
				left = val{v: left.v - right.v}
			}
		case tokMul:
			left = val{v: left.v * right.v}
		case tokDiv:
			left = val{v: left.v / right.v}
		case tokPow:
			left = val{v: math.Pow(left.v, right.v)}
		default:
			return val{}, fmt.Errorf("unexpected operator: %s", t.Text)
		}
	}

	return left, nil
}

func infixPrec(k TokenKind) (prec int, rightAssoc bool) {
	switch k {
	case tokPlus, tokMinus:
		return precAdd, false
	case tokMul, tokDiv:
		return precMul, false
	case tokPow:
		return precPow, true
	default:
		return -1, false
	}
}

func (p *parser) parsePrefix() (val, error) {
	t := p.cur()
	switch t.Kind {
	case tokPlus:
		p.pos++
		return p.parsePrefix()
	case tokMinus:
		p.pos++
		v, err := p.parsePrefix()
		if err != nil {
			return val{}, err
		}
		return val{v: -v.v, pct: v.pct}, nil
	case tokNumber:
		p.pos++
		return val{v: t.Num, pct: t.Pct}, nil
	case tokRef:
		p.pos++
		if p.refs == nil {
			return val{}, fmt.Errorf("no resolver for reference %s", t.Text)
		}
		rv, err := p.refs(t.Ref)
		if err != nil {
			return val{}, err
		}
		return val{v: rv}, nil
	case tokIdent:
		p.pos++
		fn := t.Text
		_, err := p.eat(tokLParen)
		if err != nil {
			return val{}, err
		}
		arg, err := p.parseExpr(0)
		if err != nil {
			return val{}, err
		}
		_, err = p.eat(tokRParen)
		if err != nil {
			return val{}, err
		}
		out, err := callFn(fn, arg.v)
		if err != nil {
			return val{}, err
		}
		return val{v: out}, nil
	case tokLParen:
		p.pos++
		v, err := p.parseExpr(0)
		if err != nil {
			return val{}, err
		}
		_, err = p.eat(tokRParen)
		if err != nil {
			return val{}, err
		}
		return v, nil
	default:
		return val{}, fmt.Errorf("unexpected token: %s", t.Text)
	}
}
