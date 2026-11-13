package ai

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
)

// checkNumbersGrounded is the analyst agent's check of its final response (see CompleteOptions.CheckResponse).
// It returns feedback when the response states numbers that no tool call, tool result or user message contains.
// Such numbers come from arithmetic the LLM did on query results itself, which is unreliable:
// asked for a day-over-day change, a model subtracted two query results itself and reported the wrong delta.
func checkNumbersGrounded(response string, messages []*aiv1.CompletionMessage) string {
	ungrounded := ungroundedNumbers(response, messages)
	if len(ungrounded) == 0 {
		return ""
	}
	return fmt.Sprintf(`Your response states numbers that do not appear in any tool result: %s.
Numbers you calculate yourself are often wrong, so never subtract, divide, add up or average query results yourself:
- For a change between two periods (such as day over day, week over week or month over month), call %q with "time_range" set to the later period and "comparison_time_range" set to the earlier period, without the time dimension in "dimensions", and request measures with "comparison_delta" (absolute change), "comparison_ratio" (relative change) and "comparison_value" (value in the earlier period). Report the values it returns.
- For a share of a total, request a measure with "percent_of_total".
- If no query can return a number, leave it out or say that the tools cannot compute it.
Then write your complete final response again. The user has not seen your previous response, so do not refer to it.`, strings.Join(ungrounded, ", "), QueryMetricsViewName)
}

// ungroundedNumbers returns the numbers written in a response that are not close to any number in the tool calls, tool results or user messages.
// Assistant and system messages are not sources, since the numbers to check come from the assistant.
// It ignores numbers that rarely state a measured value: small integers (counts and ranks), years, the parts of dates and times, and numbers in links.
func ungroundedNumbers(response string, messages []*aiv1.CompletionMessage) []string {
	var sources []float64
	for _, msg := range messages {
		for _, block := range msg.Content {
			switch b := block.BlockType.(type) {
			case *aiv1.ContentBlock_ToolResult:
				sources = appendContentNumbers(sources, b.ToolResult.Content)
			case *aiv1.ContentBlock_ToolCall:
				sources = appendValueNumbers(sources, b.ToolCall.Input.AsMap())
			case *aiv1.ContentBlock_Text:
				if msg.Role == string(RoleUser) {
					sources = appendTextNumbers(sources, b.Text)
				}
			}
		}
	}

	// The numbers in links are IDs, not values.
	response = markdownLinkTargetRegexp.ReplaceAllString(response, "]")
	response = urlRegexp.ReplaceAllString(response, "")

	var res []string
	for _, n := range scanNumbers(response) {
		if n.ordinal || n.isPlainCount() {
			continue
		}
		if !n.groundedIn(sources) && !slices.Contains(res, n.text) {
			res = append(res, n.text)
		}
	}
	return res
}

var (
	markdownLinkTargetRegexp = regexp.MustCompile(`\]\([^)\s]*\)`)
	urlRegexp                = regexp.MustCompile(`https?://[^\s)\]>]+`)
	// numberRegexp matches digits with the separators of common locales: 1,234.5 and 1.234,5 and 1'234 and 1 234 (with a no-break or thin space).
	// Regular spaces are not separators, so "1 234" is two numbers.
	numberRegexp = regexp.MustCompile(`\d(?:[\d.,'’\x{00A0}\x{202F}\x{2009}]*\d)?`)
)

// writtenNumber is a number written in text.
type writtenNumber struct {
	// text is the number as written, with its sign, currency symbol and suffix, for feedback to the LLM.
	text string
	// readings are the values the number may stand for: "1,234" is 1234 in English but 1.234 in German.
	readings []numberReading
	// percent is set for a percentage, which may state a ratio ("1.66%" for 0.0166) or a value that is already in percent.
	percent bool
	// unknownUnit is set when a word follows the number that may be a scale in another language (like "milliards"),
	// so the number is also matched by its significant digits, regardless of magnitude.
	unknownUnit bool
	// plain is set for an integer written without separators, decimals, percent sign or scale, like a count, rank or year.
	plain bool
	// ordinal is set for "1st", "2nd", "3rd" and "4th".
	ordinal bool
}

// numberReading is one value a written number may stand for.
type numberReading struct {
	// value is the absolute value, with any scale suffix applied.
	value float64
	// tolerance is how far a source value may be from value, for the rounding the written precision implies.
	tolerance float64
	// digits are the significant digits as written, used when the scale is unknown.
	digits string
}

// isPlainCount reports whether n is a plain integer that rarely states a measured value: a small count or rank, or a year.
func (n writtenNumber) isPlainCount() bool {
	if !n.plain || len(n.readings) != 1 {
		return false
	}
	v := n.readings[0].value
	return v < 1000 || (v >= 1900 && v <= 2100)
}

// groundedIn reports whether a source value matches n, allowing for the rounding n's written precision implies.
func (n writtenNumber) groundedIn(sources []float64) bool {
	for _, r := range n.readings {
		candidates := []numberReading{r}
		if n.percent {
			candidates = append(candidates, numberReading{value: r.value / 100, tolerance: r.tolerance / 100})
		}
		for _, c := range candidates {
			for _, s := range sources {
				if math.Abs(s-c.value) <= c.tolerance+1e-9*math.Max(s, c.value) {
					return true
				}
			}
		}
		// With fewer than three significant digits, nearly any set of sources would match.
		if n.unknownUnit && len(r.digits) >= 3 && digitsMatch(r.digits, sources) {
			return true
		}
	}
	return false
}

// digitsMatch reports whether a source value has the given significant digits, give or take one in the last digit.
func digitsMatch(digits string, sources []float64) bool {
	if len(digits) > 15 { // The precision of a float64
		digits = digits[:15]
	}
	want, err := strconv.ParseFloat(digits, 64)
	if err != nil {
		return false
	}
	for _, s := range sources {
		if s == 0 {
			continue
		}
		exp := math.Floor(math.Log10(s))
		got := s / math.Pow(10, exp-float64(len(digits))+1)
		if math.Abs(got-want) <= 1 {
			return true
		}
	}
	return false
}

// scanNumbers returns the numbers written in text.
// It skips digits that are part of a word or identifier, like "Q3", "v2" or "top_10".
func scanNumbers(text string) []writtenNumber {
	var res []writtenNumber
	for _, loc := range numberRegexp.FindAllStringIndex(text, -1) {
		start, end := loc[0], loc[1]
		if prev, _ := utf8.DecodeLastRuneInString(text[:start]); start > 0 && (unicode.IsLetter(prev) || prev == '_') {
			continue
		}
		raw := text[start:end]
		readings := parseDigits(raw)
		if len(readings) == 0 { // Not a number, like a version number "1.2.3"
			continue
		}

		// Include a sign or currency symbol in the text, for readable feedback.
		textStart := start
		for range 2 {
			prev, size := utf8.DecodeLastRuneInString(text[:textStart])
			if size == 0 || !strings.ContainsRune("+-−$€£¥₹", prev) {
				break
			}
			textStart -= size
		}

		suffix := parseNumberSuffix(text[end:])
		for i := range readings {
			readings[i].value *= suffix.scale
			readings[i].tolerance *= suffix.scale
		}
		res = append(res, writtenNumber{
			text:        text[textStart : end+suffix.length],
			readings:    readings,
			percent:     suffix.percent,
			unknownUnit: suffix.unknownUnit,
			plain:       suffix.scale == 1 && !suffix.percent && strings.IndexFunc(raw, func(r rune) bool { return !unicode.IsDigit(r) }) == -1,
			ordinal:     suffix.ordinal,
		})
	}
	return res
}

// parseDigits returns the values a run of digits and separators may stand for.
// It returns nil when the separators don't form a number, like in a version number "1.2.3".
func parseDigits(raw string) []numberReading {
	// Apostrophes and narrow spaces only group digits: 1'234 and 1 234.
	s := strings.NewReplacer("'", "", "’", "", " ", "", " ", "", " ", "").Replace(raw)

	dots, commas := strings.Count(s, "."), strings.Count(s, ",")
	switch {
	case dots == 0 && commas == 0:
		return []numberReading{integerReading(s)}
	case dots > 0 && commas > 0:
		// The last separator marks the decimals: 1,234.5 and 1.234,5.
		decimal, group := s[strings.LastIndexAny(s, ".,")], ","
		if decimal == ',' {
			group = "."
		}
		if strings.Count(s, string(decimal)) > 1 {
			return nil
		}
		s = strings.ReplaceAll(s, group, "")
		return []numberReading{decimalReading(strings.Replace(s, ",", ".", 1))}
	default:
		sep := "."
		if commas > 0 {
			sep = ","
		}
		parts := strings.Split(s, sep)
		if len(parts) > 2 {
			// A repeated separator groups thousands: 1,234,567 and 1.234.567.
			for _, p := range parts[1:] {
				if len(p) != 3 {
					return nil
				}
			}
			return []numberReading{integerReading(strings.Join(parts, ""))}
		}
		if len(parts[1]) == 3 {
			// A single separator followed by three digits groups thousands in English, but marks decimals in German.
			return []numberReading{integerReading(parts[0] + parts[1]), decimalReading(parts[0] + "." + parts[1])}
		}
		return []numberReading{decimalReading(parts[0] + "." + parts[1])}
	}
}

// integerReading reads a run of digits as an integer.
// Trailing zeros may come from rounding, so 1,077,000,000 matches values within 500,000 of it.
func integerReading(s string) numberReading {
	v, _ := strconv.ParseFloat(s, 64)
	digits := strings.TrimLeft(s, "0")
	trimmed := strings.TrimRight(digits, "0")
	return numberReading{
		value:     v,
		tolerance: 0.5 * math.Pow(10, float64(len(digits)-len(trimmed))),
		digits:    trimmed,
	}
}

// decimalReading reads digits with a decimal point.
// Models round sloppily (17.545 to 17.6 by way of 17.55), so the tolerance is one unit in the last decimal.
func decimalReading(s string) numberReading {
	v, _ := strconv.ParseFloat(s, 64)
	decimals := len(s) - strings.IndexByte(s, '.') - 1
	return numberReading{
		value:     v,
		tolerance: math.Pow(10, -float64(decimals)),
		digits:    strings.TrimLeft(strings.Replace(s, ".", "", 1), "0"),
	}
}

// numberSuffix is what follows a written number.
type numberSuffix struct {
	scale       float64
	percent     bool
	unknownUnit bool
	ordinal     bool
	// length is the number of bytes of the suffix to include in the number's text.
	length int
}

var (
	// scaleSuffixes are case-sensitive scale suffixes: "17.5M" is a million, but "15m" is likely minutes.
	scaleSuffixes = map[string]float64{"k": 1e3, "K": 1e3, "M": 1e6, "MM": 1e6, "B": 1e9, "T": 1e12}
	// scaleWords are scale words, matched case-insensitively.
	scaleWords = map[string]float64{
		"thousand": 1e3, "thousands": 1e3,
		"million": 1e6, "millions": 1e6, "mn": 1e6, "mio": 1e6, "mln": 1e6,
		"billion": 1e9, "billions": 1e9, "bn": 1e9, "bln": 1e9, "mrd": 1e9,
		"trillion": 1e12, "trillions": 1e12, "tn": 1e12,
	}
)

// parseNumberSuffix parses the suffix of a written number from the text that follows it.
func parseNumberSuffix(rest string) numberSuffix {
	res := numberSuffix{scale: 1}

	// A space may separate a number from its suffix: "1,66 %" and "17.5 million".
	afterSpace := rest
	for _, space := range []string{" ", " ", " "} {
		if strings.HasPrefix(rest, space) {
			afterSpace = rest[len(space):]
			break
		}
	}
	spaced := len(afterSpace) < len(rest)

	for _, sign := range []string{"%", "％"} {
		if strings.HasPrefix(afterSpace, sign) {
			res.percent = true
			res.length = len(rest) - len(afterSpace) + len(sign)
			return res
		}
	}

	wordLen := strings.IndexFunc(afterSpace, func(r rune) bool { return !unicode.IsLetter(r) })
	if wordLen == -1 {
		wordLen = len(afterSpace)
	}
	word := afterSpace[:wordLen]
	if word == "" {
		return res
	}
	lower := strings.ToLower(word)
	length := len(rest) - len(afterSpace) + wordLen

	switch {
	case !spaced && slices.Contains([]string{"st", "nd", "rd", "th"}, lower):
		res.ordinal = true
	case scaleSuffixes[word] != 0:
		res.scale = scaleSuffixes[word]
		res.length = length
	case scaleWords[lower] != 0:
		res.scale = scaleWords[lower]
		res.length = length
	case lower == "percent" || lower == "pct":
		res.percent = true
		res.length = length
	default:
		// A unit ("seconds"), an unrelated word ("on"), or a scale in another language ("milliards").
		res.unknownUnit = true
	}
	return res
}

// appendContentNumbers appends the numbers in a tool result, which is usually JSON.
func appendContentNumbers(sources []float64, content string) []float64 {
	var v any
	if err := json.Unmarshal([]byte(content), &v); err != nil {
		return appendTextNumbers(sources, content)
	}
	return appendValueNumbers(sources, v)
}

// appendValueNumbers appends the numbers in a decoded JSON value, including the numbers written in its strings.
func appendValueNumbers(sources []float64, v any) []float64 {
	switch v := v.(type) {
	case float64:
		sources = append(sources, math.Abs(v))
	case string:
		sources = appendTextNumbers(sources, v)
	case []any:
		for _, e := range v {
			sources = appendValueNumbers(sources, e)
		}
	case map[string]any:
		for k, e := range v {
			sources = appendTextNumbers(sources, k)
			sources = appendValueNumbers(sources, e)
		}
	}
	return sources
}

// appendTextNumbers appends every value the numbers written in text may stand for.
func appendTextNumbers(sources []float64, text string) []float64 {
	for _, n := range scanNumbers(text) {
		for _, r := range n.readings {
			sources = append(sources, r.value)
			if n.percent {
				sources = append(sources, r.value/100)
			}
		}
	}
	return sources
}
