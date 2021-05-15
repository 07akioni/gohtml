package main

import (
	"gohtml/lexer"
	"gohtml/shared"
)

func main() {
	// shared.PrettyPrint(lexer.MakeLexer("<div />").Lex())
	// shared.PrettyPrint(lexer.MakeLexer("<div aa=\"aa\" bb = \"bb\" cc= \"cc\" dd =\"dd\" ee ff/>").Lex())
	// shared.PrettyPrint(lexer.MakeLexer("<div aa=\"aa\" />").Lex())
	shared.PrettyPrint(lexer.MakeLexer("<div class=\"a\"><span b>ccc</span><hr /><span></span></div>").Lex())
}
