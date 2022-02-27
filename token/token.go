// original from https://interpreterbook.com/

package token

type TokenType string

type Token struct {
	Type    TokenType // token 的类型
	Literal string    // token 的值
}

// token 的类型
const (
	ILLEGAL = "ILLEGAL" // lex 时遇到不明的字符
	EOF     = "EOF"     // 这个是当源码到达末尾时，nextToken() 函数返回的值

	// 标识符和字面值
	IDENT  = "IDENT"  // add, foobar, x, y, ...
	INT    = "INT"    // 1343456
	STRING = "STRING" // "foobar"

	// 操作符
	ASSIGN   = "="
	PLUS     = "+"
	MINUS    = "-"
	ASTERISK = "*"
	SLASH    = "/"

	BANG = "!"

	LT = "<"
	GT = ">"

	EQ     = "=="
	NOT_EQ = "!="

	AND = "&&"
	OR  = "||"

	// 分隔符
	COMMA     = ","
	SEMICOLON = ";"
	COLON     = ":"

	// 括号
	LPAREN   = "("
	RPAREN   = ")"
	LBRACE   = "{"
	RBRACE   = "}"
	LBRACKET = "["
	RBRACKET = "]"

	// 关键字
	FUNCTION = "FUNCTION"
	LET      = "LET"

	IF     = "IF"
	ELSE   = "ELSE"
	RETURN = "RETURN"

	TRUE  = "TRUE"
	FALSE = "FALSE"
)

var keywords = map[string]TokenType{
	"fn":     FUNCTION,
	"let":    LET,
	"if":     IF,
	"else":   ELSE,
	"return": RETURN,

	"true":  TRUE,
	"false": FALSE,
}

func LookupTokenType(s string) TokenType {
	if tokenType, ok := keywords[s]; ok {
		return tokenType
	} else {
		return IDENT
	}
}
