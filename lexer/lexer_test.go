// original from https://interpreterbook.com/

package lexer

import (
	"testing"
	"toyvm/token"
)

func TestNextToken1(t *testing.T) {
	input := `=+(){},;`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.ASSIGN, "="},
		{token.PLUS, "+"},
		{token.LPAREN, "("},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RBRACE, "}"},
		{token.COMMA, ","},
		{token.SEMICOLON, ";"},
		{token.EOF, ""},
	}

	lx := New(input)

	for i, test := range tests {
		tk := lx.NextToken()

		if tk.Type != test.expectedType {
			t.Fatalf("tests [%d] - token type wrong. expected %q, actual %q",
				i, test.expectedType, tk.Type)
		}

		if tk.Literal != test.expectedLiteral {
			t.Fatalf("tests [%d] - token value wrong. expected %q, actual %q",
				i, test.expectedLiteral, tk.Literal)
		}
	}
}

func TestNextToken2(t *testing.T) {
	input := `let five = 5;
	let ten = 10;
	let add = fn(x, y) {
	x + y;
	};
	let result = add(five, ten);
	`
	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.LET, "let"},
		{token.IDENT, "five"},
		{token.ASSIGN, "="},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},
		{token.LET, "let"},
		{token.IDENT, "ten"},
		{token.ASSIGN, "="},
		{token.INT, "10"},
		{token.SEMICOLON, ";"},
		{token.LET, "let"},
		{token.IDENT, "add"},
		{token.ASSIGN, "="},
		{token.FUNCTION, "fn"},
		{token.LPAREN, "("},
		{token.IDENT, "x"},
		{token.COMMA, ","},
		{token.IDENT, "y"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.IDENT, "x"},
		{token.PLUS, "+"},
		{token.IDENT, "y"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.SEMICOLON, ";"},
		{token.LET, "let"},
		{token.IDENT, "result"},
		{token.ASSIGN, "="},
		{token.IDENT, "add"},
		{token.LPAREN, "("},
		{token.IDENT, "five"},
		{token.COMMA, ","},
		{token.IDENT, "ten"},
		{token.RPAREN, ")"},
		{token.SEMICOLON, ";"},
		{token.EOF, ""},
	}

	lx := New(input)

	for i, test := range tests {
		tk := lx.NextToken()

		if tk.Type != test.expectedType {
			t.Fatalf("tests [%d] - token type wrong. expected %q, actual %q",
				i, test.expectedType, tk.Type)
		}

		if tk.Literal != test.expectedLiteral {
			t.Fatalf("tests [%d] - token value wrong. expected %q, actual %q",
				i, test.expectedLiteral, tk.Literal)
		}
	}
}

func TestNextToken3(t *testing.T) {
	input := `let five = 5;
	let ten = 10;

	let add = fn(x, y) {
		x + y;
	};

	let result = add(five, ten);
	!-/*5;
	5 < 10 > 5;

	if (5 < 10) {
		return true;
	} else {
		return false;
	}

	10 == 10;
	10 != 9;
	`

	lx := New(input)

	for tk := lx.NextToken(); tk.Type != token.EOF; tk = lx.NextToken() {
		// uncomment to check the output
		// fmt.Println(tk)
	}
}

func TestNextToken4(t *testing.T) {
	input := `
	"foobar"
	"foo bar"
	`
	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.STRING, "foobar"},
		{token.STRING, "foo bar"},
		{token.EOF, ""},
	}

	lx := New(input)

	for i, test := range tests {
		tk := lx.NextToken()

		if tk.Type != test.expectedType {
			t.Fatalf("tests [%d] - token type wrong. expected %q, actual %q",
				i, test.expectedType, tk.Type)
		}

		if tk.Literal != test.expectedLiteral {
			t.Fatalf("tests [%d] - token value wrong. expected %q, actual %q",
				i, test.expectedLiteral, tk.Literal)
		}
	}
}

func TestNextToken5(t *testing.T) {
	input := `
	[1, 2]
	`
	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.LBRACKET, "["},
		{token.INT, "1"},
		{token.COMMA, ","},
		{token.INT, "2"},
		{token.RBRACKET, "]"},
		{token.EOF, ""},
	}

	lx := New(input)

	for i, test := range tests {
		tk := lx.NextToken()

		if tk.Type != test.expectedType {
			t.Fatalf("tests [%d] - token type wrong. expected %q, actual %q",
				i, test.expectedType, tk.Type)
		}

		if tk.Literal != test.expectedLiteral {
			t.Fatalf("tests [%d] - token value wrong. expected %q, actual %q",
				i, test.expectedLiteral, tk.Literal)
		}
	}
}

func TestNextToken6(t *testing.T) {
	input := `
	{"foo": "bar"}
	`
	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.LBRACE, "{"},
		{token.STRING, "foo"},
		{token.COLON, ":"},
		{token.STRING, "bar"},
		{token.RBRACE, "}"},
		{token.EOF, ""},
	}

	lx := New(input)

	for i, test := range tests {
		tk := lx.NextToken()

		if tk.Type != test.expectedType {
			t.Fatalf("tests [%d] - token type wrong. expected %q, actual %q",
				i, test.expectedType, tk.Type)
		}

		if tk.Literal != test.expectedLiteral {
			t.Fatalf("tests [%d] - token value wrong. expected %q, actual %q",
				i, test.expectedLiteral, tk.Literal)
		}
	}
}
