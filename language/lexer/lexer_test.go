package lexer

import (
	"reflect"
	"testing"

	"github.com/sprucehealth/graphql/language/source"
)

type Test struct {
	Body     string
	Expected interface{}
}

func createSource(body string) *source.Source {
	return source.New("GraphQL", body)
}

func TestSkipsWhiteSpace(t *testing.T) {
	tests := []Test{
		{
			Body: `

    foo

`,
			Expected: []Token{{
				Kind:  NAME,
				Start: 6,
				End:   9,
				Value: "foo",
			}},
		},
		{
			Body: `
    #comment1
    foo#comment2
`,
			Expected: []Token{
				{
					Kind:  COMMENT,
					Start: 6,
					End:   14,
					Value: "comment1",
				},
				{
					Kind:  NAME,
					Start: 19,
					End:   22,
					Value: "foo",
				},
				{
					Kind:  COMMENT,
					Start: 23,
					End:   31,
					Value: "comment2",
				},
			},
		},
		{
			Body: `,,,foo,,,`,
			Expected: []Token{{
				Kind:  NAME,
				Start: 3,
				End:   6,
				Value: "foo",
			}},
		},
	}
	for _, test := range tests {
		lex := New(source.New("", test.Body))
		var tokens []Token
		for {
			tok, err := lex.NextToken()
			if err != nil {
				t.Fatal(err)
			}
			if tok.Kind == EOF {
				break
			}
			tokens = append(tokens, tok)
		}
		if !reflect.DeepEqual(tokens, test.Expected) {
			t.Fatalf("unexpected token, expected: %v, got: %v, body: %s", test.Expected, tokens, test.Body)
		}
	}
}

func TestErrorsRespectWhitespace(t *testing.T) {
	body := `

    ?

`
	_, err := New(createSource(body)).NextToken()
	expected := "Syntax Error GraphQL (3:5) Unexpected character \"?\".\n\n2: \n3:     ?\n       ^\n4: \n"
	if err == nil {
		t.Fatalf("unexpected nil error\nexpected:\n%v\n\ngot:\n%v", expected, err)
	}
	if err.Error() != expected {
		t.Fatalf("unexpected error.\nexpected:\n%v\n\ngot:\n%v", expected, err.Error())
	}
}

func TestLexesStrings(t *testing.T) {
	tests := []Test{
		{
			Body: "\"simple\"",
			Expected: Token{
				Kind:  STRING,
				Start: 0,
				End:   8,
				Value: "simple",
			},
		},
		{
			Body: "\" white space \"",
			Expected: Token{
				Kind:  STRING,
				Start: 0,
				End:   15,
				Value: " white space ",
			},
		},
		{
			Body: "\"quote \\\"\"",
			Expected: Token{
				Kind:  STRING,
				Start: 0,
				End:   10,
				Value: `quote "`,
			},
		},
		{
			Body: "\"escaped \\n\\r\\b\\t\\f\"",
			Expected: Token{
				Kind:  STRING,
				Start: 0,
				End:   20,
				Value: "escaped \n\r\b\t\f",
			},
		},
		{
			Body: "\"slashes \\\\ \\/\"",
			Expected: Token{
				Kind:  STRING,
				Start: 0,
				End:   15,
				Value: "slashes \\ \\/",
			},
		},
		{
			Body: "\"unicode \\u1234\\u5678\\u90AB\\uCDEF\"",
			Expected: Token{
				Kind:  STRING,
				Start: 0,
				End:   34,
				Value: "unicode \u1234\u5678\u90AB\uCDEF",
			},
		},
	}
	for _, test := range tests {
		token, err := New(source.New("", test.Body)).NextToken()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(token, test.Expected) {
			t.Fatalf("unexpected token, expected: %v, got: %v", test.Expected, token)
		}
	}
}

func TestLexReportsUsefulStringErrors(t *testing.T) {
	tests := []Test{
		{
			Body: "\"no end quote",
			Expected: `Syntax Error GraphQL (1:14) Unterminated string.

1: "no end quote
                ^
`,
		},
		{
			Body: "\"multi\nline\"",
			Expected: `Syntax Error GraphQL (1:7) Unterminated string.

1: "multi
         ^
2: line"
`,
		},
		{
			Body: "\"multi\rline\"",
			Expected: `Syntax Error GraphQL (1:7) Unterminated string.

1: "multi
         ^
2: line"
`,
		},
		{
			Body: "\"multi\u2028line\"",
			Expected: `Syntax Error GraphQL (1:7) Unterminated string.

1: "multi
         ^
2: line"
`,
		},
		{
			Body: "\"multi\u2029line\"",
			Expected: `Syntax Error GraphQL (1:7) Unterminated string.

1: "multi
         ^
2: line"
`,
		},
		{
			Body: "\"bad \\z esc\"",
			Expected: `Syntax Error GraphQL (1:7) Bad character escape sequence.

1: "bad \z esc"
         ^
`,
		},
		{
			Body: "\"bad \\x esc\"",
			Expected: `Syntax Error GraphQL (1:7) Bad character escape sequence.

1: "bad \x esc"
         ^
`,
		},
		{
			Body: "\"bad \\u1 esc\"",
			Expected: `Syntax Error GraphQL (1:7) Bad character escape sequence.

1: "bad \u1 esc"
         ^
`,
		},
		{
			Body: "\"bad \\u0XX1 esc\"",
			Expected: `Syntax Error GraphQL (1:7) Bad character escape sequence.

1: "bad \u0XX1 esc"
         ^
`,
		},
		{
			Body: "\"bad \\uXXXX esc\"",
			Expected: `Syntax Error GraphQL (1:7) Bad character escape sequence.

1: "bad \uXXXX esc"
         ^
`,
		},
		{
			Body: "\"bad \\uFXXX esc\"",
			Expected: `Syntax Error GraphQL (1:7) Bad character escape sequence.

1: "bad \uFXXX esc"
         ^
`,
		},
		{
			Body: "\"bad \\uXXXF esc\"",
			Expected: `Syntax Error GraphQL (1:7) Bad character escape sequence.

1: "bad \uXXXF esc"
         ^
`,
		},
	}
	for _, test := range tests {
		_, err := New(createSource(test.Body)).NextToken()
		if err == nil {
			t.Fatalf("unexpected nil error\nexpected:\n%v\n\ngot:\n%v", test.Expected, err)
		}
		if err.Error() != test.Expected {
			t.Fatalf("unexpected error.\nexpected:\n%v\n\ngot:\n%v", test.Expected, err.Error())
		}
	}
}

func TestLexesNumbers(t *testing.T) {
	tests := []Test{
		{
			Body: "4",
			Expected: Token{
				Kind:  INT,
				Start: 0,
				End:   1,
				Value: "4",
			},
		},
		{
			Body: "4.123",
			Expected: Token{
				Kind:  FLOAT,
				Start: 0,
				End:   5,
				Value: "4.123",
			},
		},
		{
			Body: "-4",
			Expected: Token{
				Kind:  INT,
				Start: 0,
				End:   2,
				Value: "-4",
			},
		},
		{
			Body: "9",
			Expected: Token{
				Kind:  INT,
				Start: 0,
				End:   1,
				Value: "9",
			},
		},
		{
			Body: "0",
			Expected: Token{
				Kind:  INT,
				Start: 0,
				End:   1,
				Value: "0",
			},
		},
		{
			Body: "-4.123",
			Expected: Token{
				Kind:  FLOAT,
				Start: 0,
				End:   6,
				Value: "-4.123",
			},
		},
		{
			Body: "0.123",
			Expected: Token{
				Kind:  FLOAT,
				Start: 0,
				End:   5,
				Value: "0.123",
			},
		},
		{
			Body: "123e4",
			Expected: Token{
				Kind:  FLOAT,
				Start: 0,
				End:   5,
				Value: "123e4",
			},
		},
		{
			Body: "123E4",
			Expected: Token{
				Kind:  FLOAT,
				Start: 0,
				End:   5,
				Value: "123E4",
			},
		},
		{
			Body: "123e-4",
			Expected: Token{
				Kind:  FLOAT,
				Start: 0,
				End:   6,
				Value: "123e-4",
			},
		},
		{
			Body: "123e+4",
			Expected: Token{
				Kind:  FLOAT,
				Start: 0,
				End:   6,
				Value: "123e+4",
			},
		},
		{
			Body: "-1.123e4",
			Expected: Token{
				Kind:  FLOAT,
				Start: 0,
				End:   8,
				Value: "-1.123e4",
			},
		},
		{
			Body: "-1.123E4",
			Expected: Token{
				Kind:  FLOAT,
				Start: 0,
				End:   8,
				Value: "-1.123E4",
			},
		},
		{
			Body: "-1.123e-4",
			Expected: Token{
				Kind:  FLOAT,
				Start: 0,
				End:   9,
				Value: "-1.123e-4",
			},
		},
		{
			Body: "-1.123e+4",
			Expected: Token{
				Kind:  FLOAT,
				Start: 0,
				End:   9,
				Value: "-1.123e+4",
			},
		},
		{
			Body: "-1.123e4567",
			Expected: Token{
				Kind:  FLOAT,
				Start: 0,
				End:   11,
				Value: "-1.123e4567",
			},
		},
	}
	for _, test := range tests {
		token, err := New(createSource(test.Body)).NextToken()
		if err != nil {
			t.Fatalf("unexpected error: %v, test: %s", err, test)
		}
		if !reflect.DeepEqual(token, test.Expected) {
			t.Fatalf("unexpected token, expected: %v, got: %v, test: %v", test.Expected, token, test)
		}
	}
}

func TestLexReportsUsefulNumbeErrors(t *testing.T) {
	tests := []Test{
		{
			Body: "00",
			Expected: `Syntax Error GraphQL (1:2) Invalid number, unexpected digit after 0: "0".

1: 00
    ^
`,
		},
		{
			Body: "+1",
			Expected: `Syntax Error GraphQL (1:1) Unexpected character "+".

1: +1
   ^
`,
		},
		{
			Body: "1.",
			Expected: `Syntax Error GraphQL (1:3) Invalid number, expected digit but got: EOF.

1: 1.
     ^
`,
		},
		{
			Body: ".123",
			Expected: `Syntax Error GraphQL (1:1) Unexpected character ".".

1: .123
   ^
`,
		},
		{
			Body: "1.A",
			Expected: `Syntax Error GraphQL (1:3) Invalid number, expected digit but got: "A".

1: 1.A
     ^
`,
		},
		{
			Body: "-A",
			Expected: `Syntax Error GraphQL (1:2) Invalid number, expected digit but got: "A".

1: -A
    ^
`,
		},
		{
			Body: "1.0e",
			Expected: `Syntax Error GraphQL (1:5) Invalid number, expected digit but got: EOF.

1: 1.0e
       ^
`,
		},
		{
			Body: "1.0eA",
			Expected: `Syntax Error GraphQL (1:5) Invalid number, expected digit but got: "A".

1: 1.0eA
       ^
`,
		},
	}
	for _, test := range tests {
		_, err := New(createSource(test.Body)).NextToken()
		if err == nil {
			t.Fatalf("unexpected nil error\nexpected:\n%v\n\ngot:\n%v", test.Expected, err)
		}
		if err.Error() != test.Expected {
			t.Fatalf("unexpected error.\nexpected:\n%v\n\ngot:\n%v", test.Expected, err.Error())
		}
	}
}

func TestLexesPunctuation(t *testing.T) {
	tests := []Test{
		{
			Body: "!",
			Expected: Token{
				Kind:  BANG,
				Start: 0,
				End:   1,
				Value: "",
			},
		},
		{
			Body: "$",
			Expected: Token{
				Kind:  DOLLAR,
				Start: 0,
				End:   1,
				Value: "",
			},
		},
		{
			Body: "(",
			Expected: Token{
				Kind:  PAREN_L,
				Start: 0,
				End:   1,
				Value: "",
			},
		},
		{
			Body: ")",
			Expected: Token{
				Kind:  PAREN_R,
				Start: 0,
				End:   1,
				Value: "",
			},
		},
		{
			Body: "...",
			Expected: Token{
				Kind:  SPREAD,
				Start: 0,
				End:   3,
				Value: "",
			},
		},
		{
			Body: ":",
			Expected: Token{
				Kind:  COLON,
				Start: 0,
				End:   1,
				Value: "",
			},
		},
		{
			Body: "=",
			Expected: Token{
				Kind:  EQUALS,
				Start: 0,
				End:   1,
				Value: "",
			},
		},
		{
			Body: "@",
			Expected: Token{
				Kind:  AT,
				Start: 0,
				End:   1,
				Value: "",
			},
		},
		{
			Body: "[",
			Expected: Token{
				Kind:  BRACKET_L,
				Start: 0,
				End:   1,
				Value: "",
			},
		},
		{
			Body: "]",
			Expected: Token{
				Kind:  BRACKET_R,
				Start: 0,
				End:   1,
				Value: "",
			},
		},
		{
			Body: "{",
			Expected: Token{
				Kind:  BRACE_L,
				Start: 0,
				End:   1,
				Value: "",
			},
		},
		{
			Body: "|",
			Expected: Token{
				Kind:  PIPE,
				Start: 0,
				End:   1,
				Value: "",
			},
		},
		{
			Body: "}",
			Expected: Token{
				Kind:  BRACE_R,
				Start: 0,
				End:   1,
				Value: "",
			},
		},
	}
	for _, test := range tests {
		token, err := New(createSource(test.Body)).NextToken()
		if err != nil {
			t.Fatalf("unexpected error :%v, test: %v", err, test)
		}
		if !reflect.DeepEqual(token, test.Expected) {
			t.Fatalf("unexpected token, expected: %v, got: %v, test: %v", test.Expected, token, test)
		}
	}
}

func TestLexReportsUsefulUnknownCharacterError(t *testing.T) {
	tests := []Test{
		{
			Body: "..",
			Expected: `Syntax Error GraphQL (1:1) Unexpected character ".".

1: ..
   ^
`,
		},
		{
			Body: "?",
			Expected: `Syntax Error GraphQL (1:1) Unexpected character "?".

1: ?
   ^
`,
		},
		{
			Body: "\u203B",
			Expected: `Syntax Error GraphQL (1:1) Unexpected character "※".

1: ※
   ^
`,
		},
	}
	for _, test := range tests {
		_, err := New(createSource(test.Body)).NextToken()
		if err == nil {
			t.Fatalf("unexpected nil error\nexpected:\n%v\n\ngot:\n%v", test.Expected, err)
		}
		if err.Error() != test.Expected {
			t.Fatalf("unexpected error.\nexpected:\n%v\n\ngot:\n%v", test.Expected, err.Error())
		}
	}
}

func TestLexRerportsUsefulInformationForDashesInNames(t *testing.T) {
	q := "a-b"
	lexer := New(createSource(q))
	firstToken, err := lexer.NextToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	firstTokenExpected := Token{
		Kind:  NAME,
		Start: 0,
		End:   1,
		Value: "a",
	}
	if !reflect.DeepEqual(firstToken, firstTokenExpected) {
		t.Fatalf("unexpected token, expected: %v, got: %v", firstTokenExpected, firstToken)
	}
	errExpected := `Syntax Error GraphQL (1:3) Invalid number, expected digit but got: "b".

1: a-b
     ^
`
	token, err := lexer.NextToken()
	if err == nil {
		t.Fatalf("unexpected nil error: %v", err)
	}
	if err.Error() != errExpected {
		t.Fatalf("unexpected error, token:%v\nexpected:\n%v\n\ngot:\n%v", token, errExpected, err.Error())
	}
}
