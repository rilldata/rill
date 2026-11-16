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
// It ignores numbers that rarely state a measured value: small integers without a sign or currency symbol (counts and ranks), years,
// the parts of dates and times, and numbers in links.
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
	// numberRegexp matches digits with the separators of common locales: 1,234.5 and 1.234,5 and 1'234 and 1 234,5.
	// A regular space only groups digits in threes after one to three leading digits, so "Mar 11 2025" is two numbers;
	// no-break, narrow and thin spaces group digits anywhere.
	numberRegexp = regexp.MustCompile(`\d{1,3}(?: \d{3})+(?:[.,]\d+)?\b|\d(?:[\d.,'’\x{00A0}\x{202F}\x{2009}]*\d)?`)
)

// writtenNumber is a number written in text.
type writtenNumber struct {
	// text is the number as written, with its sign, currency symbol and suffix, for feedback to the LLM.
	text string
	// readings are the values the number may stand for: "1,234" is 1234 in English but 1.234 in German.
	readings []numberReading
	// sign is 1 or -1 for a number written with a sign, which states the direction of a change, and 0 otherwise.
	sign int
	// percent is set for a percentage, which may state a ratio ("1.66%" for 0.0166) or a value that is already in percent.
	percent bool
	// plain is set for an integer written without sign, currency symbol, separators, decimals, percent sign or scale, like a count, rank or year.
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
}

// isPlainCount reports whether n is a plain integer that rarely states a measured value: a small count or rank, or a year.
// Amounts with a currency symbol and signed changes are never plain, however small: "$300" and "+200" are checked.
func (n writtenNumber) isPlainCount() bool {
	if !n.plain || len(n.readings) != 1 {
		return false
	}
	v := n.readings[0].value
	return v < 1000 || (v >= 1900 && v <= 2100)
}

// groundedIn reports whether a source value matches n, allowing for the rounding n's written precision implies.
// A number written with a sign only matches a source with the same sign;
// one written without a sign matches either, as in "fell by 5%" for a ratio of -0.05.
func (n writtenNumber) groundedIn(sources []float64) bool {
	for _, r := range n.readings {
		candidates := []numberReading{r}
		if n.percent {
			candidates = append(candidates, numberReading{value: r.value / 100, tolerance: r.tolerance / 100})
		}
		for _, c := range candidates {
			for _, s := range sources {
				if n.sign == 0 {
					s = math.Abs(s)
				} else {
					s *= float64(n.sign)
				}
				if math.Abs(s-c.value) <= c.tolerance+1e-9*math.Max(math.Abs(s), c.value) {
					return true
				}
			}
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

		// Read a sign and currency symbol before the number, and include them in the text for readable feedback.
		textStart, sign, currency := start, 0, false
	prefix:
		for range 2 {
			prev, size := utf8.DecodeLastRuneInString(text[:textStart])
			before, _ := utf8.DecodeLastRuneInString(text[:textStart-size])
			switch {
			case strings.ContainsRune("$€£¥₹", prev):
				currency = true
			case strings.ContainsRune("+-−", prev) && !unicode.IsLetter(before) && !unicode.IsDigit(before):
				// A hyphen right after a letter or digit joins a range or zone, like "10-20" or "UTC-5", so it is not a sign.
				sign = 1
				if prev != '+' {
					sign = -1
				}
			default:
				break prefix
			}
			textStart -= size
		}

		suffix := parseNumberSuffix(text[end:])
		for i := range readings {
			readings[i].value *= suffix.scale
			readings[i].tolerance *= suffix.scale
		}
		res = append(res, writtenNumber{
			text:     text[textStart : end+suffix.length],
			readings: readings,
			sign:     sign,
			percent:  suffix.percent,
			plain:    sign == 0 && !currency && suffix.scale == 1 && !suffix.percent && strings.IndexFunc(raw, func(r rune) bool { return !unicode.IsDigit(r) }) == -1,
			ordinal:  suffix.ordinal,
		})
	}
	return res
}

// parseDigits returns the values a run of digits and separators may stand for.
// It returns nil when the separators don't form a number, like in a version number "1.2.3".
func parseDigits(raw string) []numberReading {
	s := digitGroupSeparators.Replace(raw)

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
	}
}

// numberSuffix is what follows a written number.
type numberSuffix struct {
	scale   float64
	percent bool
	ordinal bool
	// length is the number of bytes of the suffix to include in the number's text.
	length int
}

var (
	// digitGroupSeparators only group digits: 1'234 and 1 234 (with a regular, no-break, narrow or thin space).
	digitGroupSeparators = strings.NewReplacer("'", "", "’", "", " ", "", "\u00a0", "", "\u202f", "", "\u2009", "")
	// scaleSuffixes are case-sensitive scale suffixes: "17.5M" is a million, but "15m" is likely minutes.
	scaleSuffixes = map[string]float64{"k": 1e3, "K": 1e3, "M": 1e6, "MM": 1e6, "B": 1e9, "T": 1e12}
	// scaleWords are scale words in the languages the analyst may answer in, matched case-insensitively.
	// A number followed by any other word is read at face value, so a scale missing here causes a needless revision, not a missed error.
	// Words that differ between languages keep their English meaning: "billion" is 1e9, though it is 1e12 in French and German.
	scaleWords = map[string]float64{
		"thousand": 1e3, "thousands": 1e3, "tsd": 1e3, "tausend": 1e3, "mil": 1e3, "mille": 1e3, "mila": 1e3, "duizend": 1e3,
		"million": 1e6, "millions": 1e6, "mn": 1e6, "mio": 1e6, "mln": 1e6, "millionen": 1e6, "millón": 1e6, "millones": 1e6,
		"milione": 1e6, "milioni": 1e6, "milhão": 1e6, "milhões": 1e6, "miljoen": 1e6,
		"billion": 1e9, "billions": 1e9, "bn": 1e9, "bln": 1e9, "mrd": 1e9, "milliard": 1e9, "milliards": 1e9, "milliarde": 1e9, "milliarden": 1e9,
		"miliardo": 1e9, "miliardi": 1e9, "bilhão": 1e9, "bilhões": 1e9, "miljard": 1e9,
		"trillion": 1e12, "trillions": 1e12, "tn": 1e12, "billón": 1e12, "billones": 1e12,
		"lakh": 1e5, "lakhs": 1e5, "crore": 1e7, "crores": 1e7, "cr": 1e7,
		// Chinese, Japanese and Korean scales may run into the next word, like "3.18亿元" and "31.8억원", so the first character of a word is matched too.
		"千": 1e3, "万": 1e4, "萬": 1e4, "亿": 1e8, "億": 1e8, "兆": 1e12, "천": 1e3, "만": 1e4, "억": 1e8, "조": 1e12,
	}
)

// parseNumberSuffix parses the suffix of a written number from the text that follows it.
func parseNumberSuffix(rest string) numberSuffix {
	res := numberSuffix{scale: 1}

	// A space may separate a number from its suffix: "1,66 %" and "17.5 million".
	afterSpace := rest
	for _, space := range []string{" ", "\u00a0", "\u202f"} {
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

	word := leadingWord(afterSpace)
	if word == "" {
		return res
	}
	lower := strings.ToLower(word)
	length := len(rest) - len(afterSpace) + len(word)
	first, firstLen := utf8.DecodeRuneInString(word)

	switch {
	case !spaced && slices.Contains([]string{"st", "nd", "rd", "th"}, lower):
		res.ordinal = true
	case scaleSuffixes[word] != 0:
		res.scale = scaleSuffixes[word]
		res.length = length
	case scaleWords[lower] != 0:
		res.scale = scaleWords[lower]
		res.length = length
		// A second scale word multiplies the first: "mil millones" in Spanish and "thousand million" in British English are 1e9.
		next := afterSpace[len(word):]
		if strings.HasPrefix(next, " ") {
			nextWord := leadingWord(next[1:])
			if scale := scaleWords[strings.ToLower(nextWord)]; scale != 0 {
				res.scale *= scale
				res.length += 1 + len(nextWord)
			}
		}
	case unicode.In(first, unicode.Han, unicode.Hangul) && scaleWords[string(first)] != 0:
		res.scale = scaleWords[string(first)]
		res.length = length - len(word) + firstLen
	case lower == "percent" || lower == "pct":
		res.percent = true
		res.length = length
	}
	return res
}

// leadingWord returns the letters at the start of s.
func leadingWord(s string) string {
	n := strings.IndexFunc(s, func(r rune) bool { return !unicode.IsLetter(r) })
	if n == -1 {
		return s
	}
	return s[:n]
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
		sources = append(sources, v)
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
// A number written without a sign may state a decrease ("fell by 5%"), so it is appended with both signs.
func appendTextNumbers(sources []float64, text string) []float64 {
	for _, n := range scanNumbers(text) {
		for _, r := range n.readings {
			values := []float64{r.value}
			if n.percent {
				values = append(values, r.value/100)
			}
			for _, v := range values {
				if n.sign == 0 {
					sources = append(sources, v, -v)
				} else {
					sources = append(sources, float64(n.sign)*v)
				}
			}
		}
	}
	return sources
}
