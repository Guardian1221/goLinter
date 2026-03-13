package analyzer

import (
	"go/ast"
	"go/token"
	"go/types"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "loglint",
	Doc:  "checks log messages for style, language and sensitive data",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			pkg, method := resolveCallee(call, pass)
			if pkg == "" || method == "" {
				return true
			}

			if !isLogMethod(method) {
				return true
			}

			if len(call.Args) == 0 {
				return true
			}

			basic, ok := call.Args[0].(*ast.BasicLit)
			if !ok || basic.Kind != token.STRING {
				return true
			}

			msg, err := strconv.Unquote(basic.Value)
			if err != nil {
				return true
			}

			checkLogMessage(pass, basic, msg)
			return true
		})
	}
	return nil, nil
}

func resolveCallee(call *ast.CallExpr, pass *analysis.Pass) (pkg, method string) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return "", ""
	}

	if id, ok := sel.X.(*ast.Ident); ok {
		if obj, ok := pass.TypesInfo.Uses[id].(*types.PkgName); ok {
			path := obj.Imported().Path()
			switch path {
			case "log/slog":
				return "log/slog", sel.Sel.Name
			}
		}
	}

	if selInfo := pass.TypesInfo.Selections[sel]; selInfo != nil {
		recv := selInfo.Recv()
		if strings.Contains(recv.String(), "go.uber.org/zap.Logger") {
			return "go.uber.org/zap", sel.Sel.Name
		}
	}

	return "", ""
}

var logMethods = map[string]struct{}{
	"Debug":  {},
	"Info":   {},
	"Warn":   {},
	"Error":  {},
	"Debugf": {},
	"Infof":  {},
	"Warnf":  {},
	"Errorf": {},
}

func isLogMethod(name string) bool {
	_, ok := logMethods[name]
	return ok
}

func checkLogMessage(pass *analysis.Pass, node ast.Node, msg string) {
	if !startsWithLowercaseLetter(msg) {
		pass.Reportf(node.Pos(), "log message should start with a lowercase letter")
	}

	if hasNonEnglishLetters(msg) {
		pass.Reportf(node.Pos(), "log message should contain only English letters")
	}

	if hasSpecialSymbolsOrEmoji(msg) {
		pass.Reportf(node.Pos(), "log message should not contain special symbols or emoji")
	}

	if kw := containsSensitiveKeyword(msg); kw != "" {
		pass.Reportf(node.Pos(), "log message should not contain potentially sensitive data (%s)", kw)
	}
}

func startsWithLowercaseLetter(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) {
			continue
		}
		return unicode.IsLower(r)
	}
	return true
}

func hasNonEnglishLetters(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) && r > unicode.MaxASCII {
			return true
		}
	}
	return false
}

var allowedPunct = ".,:;!?\"'-_/()[]{}<>@ "

func hasSpecialSymbolsOrEmoji(s string) bool {
	for _, r := range s {
		if r <= unicode.MaxASCII {
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				continue
			}
			if strings.ContainsRune(allowedPunct, r) {
				continue
			}
			// Other ASCII control or unusual chars are treated as special.
			if r < 32 {
				return true
			}
			continue
		}

		if unicode.IsSymbol(r) {
			return true
		}
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && !unicode.IsSpace(r) {
			return true
		}
	}
	return false
}

var sensitiveKeywords = []string{
	"password",
	"passwd",
	"secret",
	"token",
	"apikey",
	"api_key",
	"card",
	"cvv",
	"ssn",
	"credit card",
	"authorization",
	"bearer",
	"cookie",
	"sessionid",
	"session_id",
	"phone",
	"email",
}

func containsSensitiveKeyword(s string) string {
	lower := strings.ToLower(s)
	for _, kw := range sensitiveKeywords {
		if strings.Contains(lower, kw) {
			return kw
		}
	}
	return ""
}

