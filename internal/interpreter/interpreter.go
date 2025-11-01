package interpreter

import (
	"fmt"
	"strconv"
	"strings"
)

var variables = map[string]any{}

// Run is the entry point for the interpreter.
func Run(code string) {
	rawLines := strings.Split(code, "\n")
	lines := preprocessLines(rawLines)
	executeBlock(lines, 0, len(lines))
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
			condExpr, blockStart := parseConditionBlock(lines, i)
			blockEnd := findBlockEnd(lines, i)

			if evalCondition(condExpr) {
				executeBlock(lines, blockStart+1, blockEnd)
				i = blockEnd
				// Skip else block if present
				if blockEnd+1 < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[blockEnd+1]), "else") {
					i = findBlockEnd(lines, blockEnd+1)
				}
			} else {
				blockEnd := findBlockEnd(lines, i)
				// Run else block if present
				if blockEnd+1 < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[blockEnd+1]), "else") {
					elseStart := blockEnd + 1
					elseEnd := findBlockEnd(lines, elseStart)
					executeBlock(lines, elseStart+1, elseEnd)
					i = elseEnd
				} else {
					i = blockEnd
				}
			}
			continue
		case strings.HasPrefix(line, "else"):
			continue
		default:
			fmt.Println("Unknown command:", line)
		}
	}
}

// parseConditionBlock extracts the condition and finds the start of its block
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

// findBlockEnd finds the matching '}' index
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

// handle variable declarations
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

// handle print
func handlePrint(line string) {
	inside := strings.TrimSpace(line[len("print(") : len(line)-1])
	val := evalExpr(inside)
	if val != nil {
		fmt.Println(formatValue(val))
	}
}

// evaluate condition like x > 10
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

// compareValues handles ==, >, etc
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
		return false
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

// evaluate expression including strings and + operator
func evalExpr(expr string) any {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return nil
	}

	// string literal
	if strings.HasPrefix(expr, "\"") && strings.HasSuffix(expr, "\"") {
		return strings.Trim(expr, "\"")
	}

	// concatenation with +
	if strings.Contains(expr, "+") {
		parts := splitByPlus(expr)
		var result strings.Builder
		for _, p := range parts {
			val := evalExpr(p)
			result.WriteString(fmt.Sprintf("%v", val))
		}
		return result.String()
	}

	if v, ok := variables[expr]; ok {
		return v
	}
	if f, err := strconv.ParseFloat(expr, 64); err == nil {
		return f
	}
	return expr
}

// split by + but ignore + inside quotes
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
