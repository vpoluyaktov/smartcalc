package finance

import (
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

// FinanceExpr is the top-level union of all financial expression types.
// The parser tries each alternative in order until one matches.
type FinanceExpr struct {
	Loan             *LoanExpr             `  @@`
	Mortgage         *MortgageExpr         `| @@`
	CompoundInterest *CompoundInterestExpr `| @@`
	SimpleInterest   *SimpleInterestExpr   `| @@`
	Investment       *InvestmentExpr       `| @@`
}

// LoanExpr parses: "loan $250000 at 6.5% for 30 years"
type LoanExpr struct {
	Keyword   string  `@"loan"`
	Principal *Amount `@@`
	At        string  `"at"`
	Rate      *Rate   `@@`
	For       string  `"for"`
	Term      *Term   `@@`
}

// MortgageExpr parses: "mortgage $350000 at 7% for 30 years [extra [payment] $500] [pay schedule]"
type MortgageExpr struct {
	Keyword      string  `@"mortgage"`
	Principal    *Amount `@@`
	At           string  `"at"`
	Rate         *Rate   `@@`
	For          string  `"for"`
	Term         *Term   `@@`
	ExtraPayment *Amount `( "extra" "payment"? @@ )?`
	PaySchedule  bool    `@( "pay" "schedule" )?`
}

// CompoundInterestExpr parses:
// - "compound interest $10000 at 5% for 10 years"
// - "$10000 at 5% for 10 years compounded monthly"
type CompoundInterestExpr struct {
	// Variant 1: "compound interest $10000 at 5% for 10 years [compounded monthly]"
	Keyword   string  `( @"compound" "interest" )`
	Principal *Amount `@@`
	At        string  `"at"`
	Rate      *Rate   `@@`
	For       string  `"for"`
	Term      *Term   `@@`
	Frequency string  `( "compounded" @Ident )?`
}

// SimpleInterestExpr parses: "simple interest $5000 at 3% for 2 years"
type SimpleInterestExpr struct {
	Keyword   string  `@"simple" "interest"`
	Principal *Amount `@@`
	At        string  `"at"`
	Rate      *Rate   `@@`
	For       string  `"for"`
	Term      *Term   `@@`
}

// InvestmentExpr parses: "invest $1000 at 7% for 20 years"
type InvestmentExpr struct {
	Keyword   string  `@"invest"`
	Principal *Amount `@@`
	At        string  `"at"`
	Rate      *Rate   `@@`
	For       string  `"for"`
	Term      *Term   `@@`
}

// Amount represents a monetary value like "$250000" or "250,000" or "250000"
type Amount struct {
	Dollar string `@"$"?`
	Value  string `@Number`
}

// Rate represents a percentage like "6.5%"
type Rate struct {
	Value   string `@Number`
	Percent string `@"%"`
}

// Term represents a time period like "30 years" or "10 year"
type Term struct {
	Value string `@Number`
	Unit  string `@( "years" | "year" | "months" | "month" )`
}

// financeLexer defines the tokens for financial expressions
var financeLexer = lexer.MustSimple([]lexer.SimpleRule{
	{Name: "Keyword", Pattern: `(?i)\b(loan|mortgage|compound|simple|interest|invest|at|for|extra|payment|compounded|pay|schedule|years?|months?)\b`},
	{Name: "Number", Pattern: `[0-9][0-9,]*(?:\.[0-9]+)?`},
	{Name: "Dollar", Pattern: `\$`},
	{Name: "Percent", Pattern: `%`},
	{Name: "Ident", Pattern: `[a-zA-Z][a-zA-Z0-9]*`},
	{Name: "Whitespace", Pattern: `\s+`},
})

// financeParser is the participle parser for financial expressions
var financeParser = participle.MustBuild[FinanceExpr](
	participle.Lexer(financeLexer),
	participle.CaseInsensitive("Keyword", "Ident"),
	participle.Elide("Whitespace"),
)

// Parse attempts to parse a financial expression using the grammar.
// Returns the parsed expression or an error if parsing fails.
func Parse(expr string) (*FinanceExpr, error) {
	return financeParser.ParseString("", expr)
}
