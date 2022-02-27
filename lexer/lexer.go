// original from https://interpreterbook.com/

package lexer

import "toyvm/token"

type Lexer struct {
	input        string
	position     int  // 当前字符的位置
	readPosition int  // 输入字符串的读取位置（即当前字符的下一个字符的位置）
	ch           byte // 当前字符（只支持 ascii）
}

func New(input string) *Lexer {
	lx := &Lexer{input: input}
	lx.readChar()
	return lx
}

func (lx *Lexer) readChar() {
	if lx.readPosition >= len(lx.input) {
		lx.ch = 0
	} else {
		lx.ch = lx.input[lx.readPosition]
	}

	// 移动光标到下一个字符
	lx.position = lx.readPosition
	lx.readPosition += 1
}

func (lx *Lexer) NextToken() token.Token {
	var tk token.Token

	for lx.skipComment() || lx.skipWhitespace() {
		//
	}

	switch lx.ch {
	case '=':
		if lx.peekChar() == '=' {
			lx.readChar() // 消耗下一个字符
			tk = token.Token{Type: token.EQ, Literal: "=="}

		} else {
			tk = newToken(token.ASSIGN, lx.ch)
		}

	case '!':
		if lx.peekChar() == '=' {
			lx.readChar() // 消耗下一个字符
			tk = token.Token{Type: token.NOT_EQ, Literal: "!="}

		} else {
			tk = newToken(token.BANG, lx.ch)
		}

	case '+':
		tk = newToken(token.PLUS, lx.ch)
	case '-':
		tk = newToken(token.MINUS, lx.ch)
	case '/':
		tk = newToken(token.SLASH, lx.ch)
	case '*':
		tk = newToken(token.ASTERISK, lx.ch)

	case '<':
		tk = newToken(token.LT, lx.ch)
	case '>':
		tk = newToken(token.GT, lx.ch)

	case ';':
		tk = newToken(token.SEMICOLON, lx.ch)
	case ',':
		tk = newToken(token.COMMA, lx.ch)
	case ':':
		tk = newToken(token.COLON, lx.ch)

	case '(':
		tk = newToken(token.LPAREN, lx.ch)
	case ')':
		tk = newToken(token.RPAREN, lx.ch)

	case '{':
		tk = newToken(token.LBRACE, lx.ch)
	case '}':
		tk = newToken(token.RBRACE, lx.ch)

	case '[':
		tk = newToken(token.LBRACKET, lx.ch)
	case ']':
		tk = newToken(token.RBRACKET, lx.ch)

	case '&':
		if lx.peekChar() == '&' {
			lx.readChar()
			tk = token.Token{Type: token.AND, Literal: "&&"}
		} else {
			tk = newToken(token.ILLEGAL, lx.ch) // 不明字符
		}

	case '|':
		if lx.peekChar() == '|' {
			lx.readChar()
			tk = token.Token{Type: token.OR, Literal: "||"}
		} else {
			tk = newToken(token.ILLEGAL, lx.ch) // 不明字符
		}

	case '"':
		tk.Type = token.STRING
		tk.Literal = lx.readString()

	case 0:
		// 到达文件末尾。
		// 无法通过调用 newToken() 函数来构造 Literal 值为空字符串的 Token 对象，
		// 所以手动指定 tk 的值。
		tk = token.Token{Type: token.EOF, Literal: ""}

	default:
		if isAlphabet(lx.ch) {
			s := lx.readIdentifier()

			tk = token.Token{Type: token.LookupTokenType(s), Literal: s}
			return tk // 跳过后面的语句，因为 readIdentifier() 已经读了下一个字符

		} else if isDigit(lx.ch) {
			s := lx.readNumber()

			tk = token.Token{Type: token.INT, Literal: s}
			return tk // 跳过后面的语句，因为 readNumber() 已经读了下一个字符

		} else {
			tk = newToken(token.ILLEGAL, lx.ch) // 不明字符
		}
	}

	lx.readChar() // 读下一个字符
	return tk
}

func (lx *Lexer) peekChar() byte {
	if lx.readPosition >= len(lx.input) {
		return 0
	} else {
		return lx.input[lx.readPosition]
	}
}

func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{
		Type:    tokenType,
		Literal: string(ch),
	}
}

func (lx *Lexer) skipWhitespace() bool {
	var found = false
	for lx.ch == ' ' || lx.ch == '\t' || lx.ch == '\n' || lx.ch == '\r' {
		found = true
		lx.readChar()
	}
	return found
}

func (lx *Lexer) skipComment() bool {
	var found = false
	if lx.ch == '/' && lx.peekChar() == '/' {
		found = true
		for !(lx.ch == '\n' || lx.ch == '\r' || lx.ch == 0) {
			lx.readChar()
		}
	}
	return found
}

func (lx *Lexer) readIdentifier() string {
	startPosition := lx.position
	for isLetter(lx.ch) {
		lx.readChar() // 读下一个字符
	}

	// 返回从 startPosition 到 lx.position 之间的字符
	return lx.input[startPosition:lx.position]
}

func (lx *Lexer) readNumber() string { // 以字符串的形式返回数字
	startPosition := lx.position
	for isDigit(lx.ch) {
		lx.readChar() // 读下一个字符
	}

	// 返回从 startPosition 到 lx.position 之间的字符
	return lx.input[startPosition:lx.position]
}

func (lx *Lexer) readString() string { // 返回的字符串值不包含前后双引号
	startPosition := lx.position
	for {
		lx.readChar() // 读下一个字符
		if lx.ch == '"' || lx.ch == 0 {
			break
		}
	}

	// 返回从 startPosition + 1 到 lx.position 之间的字符
	return lx.input[startPosition+1 : lx.position]
}

func isAlphabet(ch byte) bool {
	return ch >= 'a' && ch <= 'z' ||
		ch >= 'A' && ch <= 'Z' ||
		ch == '_'
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isLetter(ch byte) bool {
	return isAlphabet(ch) || isDigit(ch)
}
