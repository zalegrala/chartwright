package stamp

import (
	"strings"
	"testing"

	"github.com/zalegrala/chartwright/interchange"
)

// #3: a manifest value that resembles the old guessable sentinel must survive
// untouched, while the real hole is still substituted. Under the old fixed
// "HOLESENTINEL0END" token this collided and corrupted the literal value.
func TestRenderResourceSentinelNotGuessable(t *testing.T) {
	r := interchange.Resource{
		File: "templates/web/configmap.yaml",
		Manifest: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"data": map[string]interface{}{
				"note":     "HOLESENTINEL0END", // literal value resembling the old token
				"replicas": float64(0),
			},
		},
		Holes: []interchange.Hole{
			{Path: "web.replicas", Pointer: "/data/replicas", Default: float64(3)},
		},
	}
	out, err := renderResource(r)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "note: HOLESENTINEL0END") {
		t.Errorf("literal value resembling a sentinel was corrupted:\n%s", out)
	}
	if !strings.Contains(out, "replicas: {{ .Values.web.replicas | default 3 }}") {
		t.Errorf("real hole not substituted:\n%s", out)
	}
}

// #4: quote + required produce valid Helm (required piped into quote).
func TestInlineExprQuoteRequired(t *testing.T) {
	h := interchange.Hole{Path: "web.name", Required: true, Render: "quote"}
	got := inlineExpr(h)
	want := `{{ required "web.name is required" .Values.web.name | quote }}`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// #5: a block hole whose YAML key contains a colon keeps the key verbatim.
func TestSubstituteBlockPreservesColonKey(t *testing.T) {
	yamlText := "spec:\n  \"a:b\": CWTOKEN\n"
	holes := []interchange.Hole{{
		Path:    "web.thing",
		Pointer: "/spec/a:b",
		Default: map[string]interface{}{"x": "1"},
	}}
	tokens := map[int]string{0: "CWTOKEN"}
	out, err := substitute(yamlText, holes, tokens)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "  \"a:b\":") {
		t.Errorf("colon-containing key not preserved verbatim:\n%s", out)
	}
	if !strings.Contains(out, "{{- toYaml .Values.web.thing | nindent 4 }}") {
		t.Errorf("block not rendered:\n%s", out)
	}
}

// #6: a hole whose path equals a component gate is rejected.
func TestBuildValuesRejectsGateShadow(t *testing.T) {
	doc := interchange.Document{
		Components: map[string]interchange.Component{"web": {Enabled: true}},
		Resources: []interchange.Resource{
			{Holes: []interchange.Hole{{Path: "web.enabled", Default: false}}},
		},
	}
	if _, err := buildValues(doc); err == nil {
		t.Fatal("expected error when a hole shadows a component gate")
	}
}
