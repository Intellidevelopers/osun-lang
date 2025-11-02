package interpreter

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"net/http"
	"github.com/intellidevelopers/osun-lang/internal/runtime"
)

var variables = map[string]any{}

// Run is the entry point for the interpreter.
func Run(code string) {
	rawLines := strings.Split(code, "\n")
	lines := preprocessLines(rawLines)
	executeBlock(lines, 0, len(lines))
}



func SetVariable(name string, value interface{}) {
	variables[name] = value
}

// preprocessLines splits "} else {" and trims empty lines but preserves structure.
func preprocessLines(raw []string) []string {
	var out []string
	for _, l := range raw {
		line := strings.TrimSpace(l)
		if line == "" {
			continue
		}
		if strings.Contains(line, "} else {") {
			parts := strings.Split(line, "} else {")
			left := strings.TrimSpace(parts[0])
			if left != "" {
				out = append(out, left+"}")
			} else {
				out = append(out, "}")
			}
			out = append(out, "else {")
			if len(parts) > 1 && strings.TrimSpace(parts[1]) != "" {
				out = append(out, strings.TrimSpace(parts[1]))
			}
			continue
		}
		out = append(out, line)
	}
	return out
}

// interpreter/interpreter.go
func GetVariable(name string) (any, bool) {
	val, ok := variables[name]
	return val, ok
}


// executeBlock runs lines[start:end)
func executeBlock(lines []string, start, end int) {
	for i := start; i < end; i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		if line == "{" || line == "}" {
			continue
		}

		switch {
		case strings.HasPrefix(line, "let "):
			handleLet(line)
		case strings.HasPrefix(line, "print(") && strings.HasSuffix(line, ")"):
			handlePrint(line)
		case strings.HasPrefix(line, "if "):
			i = handleIfElse(lines, i)
		case strings.HasPrefix(line, "server.Handle"):
			handleServerHandle(line)
		case strings.HasPrefix(line, "server.Listen"):
			handleServerListen(line)
		default:
			if err := handleBuiltin(line); err != nil {
				fmt.Println("❌ Unknown command:", line)
			}
		}
	}
}


// -------------------- Builtin Execution ------------------------

func handleBuiltin(line string) error {
	if !strings.Contains(line, "(") || !strings.HasSuffix(line, ")") {
		return fmt.Errorf("not a function call")
	}

	fnPath := line[:strings.Index(line, "(")]
	argsStr := line[strings.Index(line, "(")+1 : len(line)-1]
	fnPath = strings.TrimSpace(fnPath)
	args := parseArgs(argsStr)

	parts := strings.Split(fnPath, ".")
	if len(parts) == 1 {
		sym := runtime.GetSymbol(parts[0])
		if sym == nil {
			return fmt.Errorf("symbol not found: %s", parts[0])
		}
		callFunction(sym, args)
		return nil
	} else if len(parts) == 2 {
		group := runtime.GetSymbol(parts[0])
		if groupMap, ok := group.(map[string]interface{}); ok {
			fn := groupMap[parts[1]]
			if fn == nil {
				return fmt.Errorf("method not found: %s", parts[1])
			}
			callFunction(fn, args)
			return nil
		}
	}
	return fmt.Errorf("invalid builtin path")
}

func parseArgs(argStr string) []interface{} {
	if strings.TrimSpace(argStr) == "" {
		return []interface{}{}
	}
	argsRaw := splitArgs(argStr)
	var args []interface{}
	for _, a := range argsRaw {
		args = append(args, evalExpr(a))
	}
	return args
}

func splitArgs(s string) []string {
	var args []string
	var cur strings.Builder
	inQuotes := false
	for _, ch := range s {
		if ch == '"' {
			inQuotes = !inQuotes
		}
		if ch == ',' && !inQuotes {
			args = append(args, strings.TrimSpace(cur.String()))
			cur.Reset()
			continue
		}
		cur.WriteRune(ch)
	}
	if cur.Len() > 0 {
		args = append(args, strings.TrimSpace(cur.String()))
	}
	return args
}

func callFunction(fn interface{}, args []interface{}) {
	fv := reflect.ValueOf(fn)
	if fv.Kind() != reflect.Func {
		fmt.Println("❌ not a function:", fn)
		return
	}

	if len(args) != fv.Type().NumIn() {
		fmt.Println("⚠️ argument mismatch for function")
	}

	in := make([]reflect.Value, len(args))
	for i, arg := range args {
		in[i] = reflect.ValueOf(arg)
	}

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("❌ runtime error:", r)
		}
	}()

	fv.Call(in)
}

// -------------------- IF / ELSE HANDLING ------------------------

func handleIfElse(lines []string, i int) int {
	condExpr, blockStart := parseConditionBlock(lines, i)
	blockEnd := findBlockEnd(lines, i)

	if evalCondition(condExpr) {
		executeBlock(lines, blockStart+1, blockEnd)
		i = blockEnd
		if blockEnd+1 < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[blockEnd+1]), "else") {
			i = findBlockEnd(lines, blockEnd+1)
		}
	} else {
		blockEnd := findBlockEnd(lines, i)
		if blockEnd+1 < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[blockEnd+1]), "else") {
			elseStart := blockEnd + 1
			elseEnd := findBlockEnd(lines, elseStart)
			executeBlock(lines, elseStart+1, elseEnd)
			i = elseEnd
		} else {
			i = blockEnd
		}
	}
	return i
}

// -------------------- Core Evaluators ------------------------

func handleLet(line string) {
	rest := strings.TrimSpace(strings.TrimPrefix(line, "let"))
	parts := strings.SplitN(rest, "=", 2)
	if len(parts) != 2 {
		fmt.Println("Syntax error in let:", line)
		return
	}
	name := strings.TrimSpace(parts[0])
	valueExpr := strings.TrimSpace(parts[1])
	val := evalExpr(valueExpr)
	variables[name] = val
}

func handlePrint(line string) {
	inside := strings.TrimSpace(line[len("print(") : len(line)-1])
	val := evalExpr(inside)
	if val != nil {
		fmt.Println(formatValue(val))
	}
}

// -------------------- HTTP SERVER HANDLERS ------------------------

func handleServerHandle(line string) {
	// Example: server.Handle("GET", "/hello", func(){ print("hi") })
	parts := strings.SplitN(line, "(", 2)
	argsStr := strings.TrimSuffix(parts[1], ")")
	args := parseArgs(argsStr)
	if len(args) < 3 {
		fmt.Println("❌ server.Handle requires 3 arguments")
		return
	}

	serverVar, ok := variables["server"]
	if !ok {
		fmt.Println("❌ server variable not found")
		return
	}
	server, ok := serverVar.(*runtime.OsunServer)
	if !ok {
		fmt.Println("❌ server variable invalid")
		return
	}

	method, ok1 := args[0].(string)
	path, ok2 := args[1].(string)
	handlerFunc, ok3 := args[2].(func())
	if !ok1 || !ok2 || !ok3 {
		fmt.Println("❌ invalid arguments for server.Handle")
		return
	}

	// Wrap the interpreter func into http.HandlerFunc
	server.Handle(method, path, func(w http.ResponseWriter, r *http.Request) {
		// This executes the interpreter-level closure
		handlerFunc()
	})
}


func handleServerListen(line string) {
	serverVar, ok := variables["server"]
	if !ok {
		fmt.Println("❌ server variable not found")
		return
	}
	server, ok := serverVar.(*runtime.OsunServer)
	if !ok {
		fmt.Println("❌ server variable invalid")
		return
	}

	server.Listen()
}

// -------------------- Expression Evaluators ------------------------

func evalCondition(expr string) bool {
	expr = strings.TrimSpace(expr)
	ops := []string{">=", "<=", "==", "!=", ">", "<"}
	for _, op := range ops {
		if strings.Contains(expr, op) {
			parts := strings.SplitN(expr, op, 2)
			left := strings.TrimSpace(parts[0])
			right := strings.TrimSpace(parts[1])
			lv := evalExpr(left)
			rv := evalExpr(right)
			return compareValues(lv, rv, op)
		}
	}
	v := evalExpr(expr)
	switch t := v.(type) {
	case bool:
		return t
	case float64:
		return t != 0
	case string:
		return t != ""
	default:
		return false
	}
}

func evalExpr(expr string) any {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return nil
	}

	if strings.HasPrefix(expr, "\"") && strings.HasSuffix(expr, "\"") {
		return strings.Trim(expr, "\"")
	}

	if v, ok := variables[expr]; ok {
		return v
	}

	if strings.Contains(expr, "+") {
		parts := splitByPlus(expr)
		var result strings.Builder
		for _, p := range parts {
			val := evalExpr(p)
			result.WriteString(fmt.Sprintf("%v", val))
		}
		return result.String()
	}

	if f, err := strconv.ParseFloat(expr, 64); err == nil {
		return f
	}

	return expr
}

func splitByPlus(expr string) []string {
	var parts []string
	var cur strings.Builder
	inQuotes := false
	for _, ch := range expr {
		if ch == '"' {
			inQuotes = !inQuotes
		}
		if ch == '+' && !inQuotes {
			parts = append(parts, strings.TrimSpace(cur.String()))
			cur.Reset()
			continue
		}
		cur.WriteRune(ch)
	}
	if cur.Len() > 0 {
		parts = append(parts, strings.TrimSpace(cur.String()))
	}
	return parts
}

// -------------------- Utilities ------------------------

func parseConditionBlock(lines []string, idx int) (string, int) {
	line := strings.TrimSpace(lines[idx])
	if strings.Contains(line, "{") {
		start := strings.Index(line, "if") + 2
		open := strings.Index(line, "{")
		return strings.TrimSpace(line[start:open]), idx
	}
	cond := strings.TrimSpace(strings.TrimPrefix(line, "if"))
	return cond, idx + 1
}

func findBlockEnd(lines []string, start int) int {
	depth := 0
	for i := start; i < len(lines); i++ {
		for _, ch := range lines[i] {
			if ch == '{' {
				depth++
			} else if ch == '}' {
				depth--
				if depth == 0 {
					return i
				}
			}
		}
	}
	return len(lines) - 1
}

func compareValues(a, b any, op string) bool {
	af, aok := toFloat(a)
	bf, bok := toFloat(b)
	if aok && bok {
		switch op {
		case ">":
			return af > bf
		case "<":
			return af < bf
		case ">=":
			return af >= bf
		case "<=":
			return af <= bf
		case "==":
			return af == bf
		case "!=":
			return af != bf
		}
	}
	as := fmt.Sprintf("%v", a)
	bs := fmt.Sprintf("%v", b)
	switch op {
	case "==":
		return as == bs
	case "!=":
		return as != bs
	default:
		return false
	}
}

func toFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case string:
		if f, err := strconv.ParseFloat(n, 64); err == nil {
			return f, true
		}
	}
	return 0, false
}

func formatValue(v any) string {
	switch t := v.(type) {
	case nil:
		return "<nil>"
	case float64:
		if t == float64(int64(t)) {
			return fmt.Sprintf("%d", int64(t))
		}
		return fmt.Sprintf("%v", t)
	default:
		return fmt.Sprintf("%v", t)
	}
}
