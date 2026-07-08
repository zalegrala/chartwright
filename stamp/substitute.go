package stamp

import (
	"fmt"
	"strings"

	"github.com/zalegrala/helmitis/interchange"
)

// substitute replaces each hole's sentinel token in the marshaled YAML text
// with the Helm expression for that hole's render mode.
func substitute(yamlText string, holes []interchange.Hole, tokens map[int]string) (string, error) {
	lines := strings.Split(yamlText, "\n")
	for i, h := range holes {
		tok := tokens[i]
		switch renderMode(h) {
		case "scalar", "quote":
			replaceTokenInline(lines, tok, inlineExpr(h))
		case "raw":
			replaceTokenInline(lines, tok, h.Raw)
		case "block", "with":
			var err error
			lines, err = replaceTokenBlock(lines, tok, h)
			if err != nil {
				return "", err
			}
		default:
			return "", fmt.Errorf("hole %q: unknown render mode %q", h.Path, renderMode(h))
		}
	}
	return strings.Join(lines, "\n"), nil
}

func replaceTokenInline(lines []string, tok, expr string) {
	for i, line := range lines {
		if strings.Contains(line, tok) {
			lines[i] = strings.Replace(line, tok, expr, 1)
			return
		}
	}
}

func replaceTokenBlock(lines []string, tok string, h interchange.Hole) ([]string, error) {
	return lines, fmt.Errorf("block substitution not implemented yet")
}
