package llm

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	maxReasonChars      = 300
	maxExplanationChars = 900
	maxActionChars      = 300
	maxExplanationSents = 4
)

// A sentence ends with . ! or ? followed by whitespace or the end. A dot inside "2.825" does not count.
var sentenceEnd = regexp.MustCompile(`[.!?]+(\s|$)`)

func sentences(s string) int { return len(sentenceEnd.FindAllStringIndex(strings.TrimSpace(s), -1)) }

// parseAndValidate turns the model's answer into a Rewrite or explains why it was rejected.
func parseAndValidate(content string, allowed *numberSet) (Rewrite, error) {
	// Some models wrap the JSON in a fence or add a preamble.
	if i, j := strings.Index(content, "{"), strings.LastIndex(content, "}"); i >= 0 && j > i {
		content = content[i : j+1]
	}
	var rw Rewrite
	if err := json.Unmarshal([]byte(content), &rw); err != nil {
		return Rewrite{}, fmt.Errorf("answer is not valid JSON: %w", err)
	}
	rw.Reason = strings.TrimSpace(rw.Reason)
	rw.Explanation = strings.TrimSpace(rw.Explanation)
	rw.RecommendedAction = strings.TrimSpace(rw.RecommendedAction)
	rw.Reason, rw.Explanation, rw.RecommendedAction = endWithPeriod(rw.Reason), endWithPeriod(rw.Explanation), endWithPeriod(rw.RecommendedAction)

	switch {
	case rw.Reason == "" || rw.Explanation == "" || rw.RecommendedAction == "":
		return Rewrite{}, fmt.Errorf("answer is missing reason, explanation or recommended_action")
	case sentences(rw.Reason) > 1 || utf8.RuneCountInString(rw.Reason) > maxReasonChars:
		return Rewrite{}, fmt.Errorf("reason must be a single short sentence")
	case sentences(rw.Explanation) > maxExplanationSents || utf8.RuneCountInString(rw.Explanation) > maxExplanationChars:
		return Rewrite{}, fmt.Errorf("explanation is longer than %d sentences", maxExplanationSents)
	case utf8.RuneCountInString(rw.RecommendedAction) > maxActionChars:
		return Rewrite{}, fmt.Errorf("recommended_action is too long")
	}
	text := rw.Reason + " " + rw.Explanation + " " + rw.RecommendedAction
	cited := extractNumbers(text)
	if len(cited) == 0 {
		return Rewrite{}, fmt.Errorf("answer cites no numbers")
	}
	for _, n := range cited {
		if !allowed.has(n) {
			return Rewrite{}, fmt.Errorf("answer cites %q, which is not in the evidence", n.text)
		}
	}
	return rw, nil
}

// endWithPeriod adds the final period models often forget. Empty text stays empty so the
// missing-field check still catches it.
func endWithPeriod(s string) string {
	if s == "" || strings.HasSuffix(s, ".") || strings.HasSuffix(s, "!") || strings.HasSuffix(s, "?") {
		return s
	}
	return s + "."
}

// ---- number grounding ----

// value is one reading of a numeric token, with the number of decimals it was written with.
type value struct {
	v        float64
	decimals int
}

// number is a numeric token found in text. Formats differ (2.825 is 2825 in es-CO and 2.825 in
// en-US), so a token carries every valid reading of it.
type number struct {
	text     string
	variants []value
}

var (
	numberRe = regexp.MustCompile(`\d+(?:[.,]\d+)*`)
	esFormat = regexp.MustCompile(`^\d{1,3}(\.\d{3})*(,\d+)?$|^\d+(,\d+)?$`)  // 1.234,5 or 1234,5
	enFormat = regexp.MustCompile(`^\d{1,3}(,\d{3})*(\.\d+)?$|^\d+(\.\d+)?$`) // 1,234.5 or 1234.5
)

func extractNumbers(s string) []number {
	var out []number
	for _, tok := range numberRe.FindAllString(s, -1) {
		var vs []value
		if esFormat.MatchString(tok) {
			if v, err := strconv.ParseFloat(strings.ReplaceAll(strings.ReplaceAll(tok, ".", ""), ",", "."), 64); err == nil {
				vs = append(vs, value{v, decimalsAfter(tok, ',')})
			}
		}
		if enFormat.MatchString(tok) {
			if v, err := strconv.ParseFloat(strings.ReplaceAll(tok, ",", ""), 64); err == nil {
				vs = append(vs, value{v, decimalsAfter(tok, '.')})
			}
		}
		if len(vs) > 0 {
			out = append(out, number{tok, vs})
		}
	}
	return out
}

// decimalsAfter counts the digits after the last sep, or 0 if there is none.
func decimalsAfter(tok string, sep byte) int {
	if i := strings.LastIndexByte(tok, sep); i >= 0 {
		return len(tok) - i - 1
	}
	return 0
}

// numberSet holds the values the model may cite: every number in the evidence, in the forms an
// operator would write them (absolute value, percentage of a fraction, parts of a timestamp).
type numberSet struct{ vals []float64 }

func (s *numberSet) add(v float64) {
	v = math.Abs(v)
	s.vals = append(s.vals, v)
	if v <= 1 {
		s.vals = append(s.vals, v*100) // 0,94 may be written as 94%
	}
}

// has accepts a number if it equals an allowed value up to the rounding its own precision implies
// (111 for 110,5; 0,7 for 0,74), or is a small whole number (dates, hours, counts), which are too
// generic to be worth policing.
func (s *numberSet) has(n number) bool {
	for _, x := range n.variants {
		if x.v == math.Trunc(x.v) && x.v <= 31 {
			return true
		}
		tol := 0.5*math.Pow(10, -float64(x.decimals)) + 1e-9
		for _, a := range s.vals {
			if math.Abs(x.v-a) <= tol {
				return true
			}
		}
	}
	return false
}

// numbersIn collects every number of v (any JSON-encodable value): numeric fields, numbers inside
// strings, and the day/hour/minute of timestamps.
func numbersIn(v any) (*numberSet, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var tree any
	if err := json.Unmarshal(b, &tree); err != nil {
		return nil, err
	}
	set := &numberSet{}
	var walk func(any)
	walk = func(x any) {
		switch t := x.(type) {
		case float64:
			set.add(t)
		case string:
			if ts, err := time.Parse(time.RFC3339, t); err == nil {
				set.add(float64(ts.Year()))
				set.add(float64(ts.Month()))
				set.add(float64(ts.Day()))
				set.add(float64(ts.Hour()))
				set.add(float64(ts.Minute()))
				return
			}
			for _, n := range extractNumbers(t) {
				for _, x := range n.variants {
					set.add(x.v)
				}
			}
		case []any:
			for _, e := range t {
				walk(e)
			}
		case map[string]any:
			for _, e := range t {
				walk(e)
			}
		}
	}
	walk(tree)
	return set, nil
}
