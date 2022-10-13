package pure

import (
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

var sqlLexer = lexer.MustSimple([]lexer.SimpleRule{
	{Name: `Keyword`, Pattern: `(?i)\b(CREATE|SOURCE|WITH)\b`},
	{Name: `Ident`, Pattern: `[a-zA-Z_][a-zA-Z0-9_]*`},
	{Name: `Number`, Pattern: `[-+]?\d*\.?\d+([eE][-+]?\d+)?`},
	{Name: `String`, Pattern: `'[^']*'|"[^"]*"`},
	{Name: `Operators`, Pattern: `<>|!=|<=|>=|[-+*/%,.()=<>]`},
	{Name: "whitespace", Pattern: `\s+`},
})

var sqlParser = participle.MustBuild[Stmt](
	participle.Lexer(sqlLexer),
	participle.Unquote("String"),
	participle.CaseInsensitive("Keyword"),
	// participle.Elide("Comment"),
	// participle.UseLookahead(2),
)

type Stmt struct {
	CreateSource *CreateSource `parser:"@@"`
}

type CreateSource struct {
	Name string `parser:"'CREATE' 'SOURCE' @Ident"`
	With With   `parser:"'WITH' (('(' @@ ')')|@@)"`
}

type With struct {
	Properties []*Property `parser:"(@@ ','?)+"`
}

type Property struct {
	Key   string `parser:"(@Ident | @String)"`
	Value Value  `parser:"'=' @@"`
}

type Value struct {
	Number  *float64 `parser:"( @Number"`
	String  *string  `parser:"| @String"`
	Boolean *Boolean `parser:"| @('TRUE' | 'FALSE')"`
	Null    bool     `parser:"| @'NULL' )"`
}

type Boolean bool

func (b *Boolean) Capture(values []string) error {
	*b = values[0] == "TRUE"
	return nil
}

func Parse(sql string) (*Stmt, error) {
	stmt, err := sqlParser.ParseString("", sql)
	if err != nil {
		return nil, err
	}
	err = visit(stmt)
	if err != nil {
		return nil, err
	}
	return stmt, nil
}

func visit(stmt *Stmt) error {
	if stmt.CreateSource != nil {
		return visitCreateSource(stmt.CreateSource)
	}
	return nil
}

func visitCreateSource(cs *CreateSource) error {
	for _, prop := range cs.With.Properties {
		prop.Key = strings.Trim(prop.Key, "'\"")
	}
	return nil
}
