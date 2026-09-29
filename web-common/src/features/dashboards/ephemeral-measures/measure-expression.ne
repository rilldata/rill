# Grammar for ephemeral measure expressions.
#
# Mirrors the server-side grammar in runtime/metricsview/measure_expression.go,
# which is the authority: measure references (bare or double-quoted
# identifiers), numeric literals, NULL, unary minus, the binary operators
# + - * / %, parentheses and function calls. Function names, arities, reserved
# words and nesting limits are validated after parsing in expression-parser.ts.
#
# Regenerate the compiled grammar with:
#   npm run build-measure-expression-grammar -w web-common

@preprocessor esmodule

main -> _ additive _ {% ([, expr]) => expr %}

additive -> additive _ [+\-] _ multiplicative {% ([left, , op, , right], pos) => ({ type: "binary", op, left, right, pos }) %}
          | multiplicative                     {% id %}

multiplicative -> multiplicative _ [*/%] _ unary {% ([left, , op, , right], pos) => ({ type: "binary", op, left, right, pos }) %}
                | unary                          {% id %}

unary -> "-" _ unary {% ([, , expr], pos) => ({ type: "unary", expr, pos }) %}
       | primary     {% id %}

primary -> number                       {% ([literal], pos) => ({ type: "literal", literal, pos }) %}
         | quoted_ident                 {% ([name], pos) => ({ type: "ref", name, pos }) %}
         | ident _ "(" _ args:? _ ")"   {% ([name, , , , args], pos) => ({ type: "func", name, args: args ?? [], pos }) %}
         | ident                        {% ([name], pos) => ({ type: "word", name, pos }) %}
         | "(" _ additive _ ")"         {% ([, , expr], pos) => ({ type: "paren", expr, pos }) %}

args -> additive (_ "," _ additive):* {% ([first, rest]) => [first, ...rest.map(([, , , arg]) => arg)] %}

ident -> [A-Za-z_] [A-Za-z0-9_]:* {% ([first, rest]) => first + rest.join("") %}

# Double-quoted identifier; "" escapes a quote, matching ANSI SQL.
quoted_ident -> "\"" qchar:+ "\"" {% ([, chars]) => chars.join("") %}
qchar -> [^"]   {% id %}
       | "\"\"" {% () => '"' %}

number -> digits ("." digits:?):? exponent:? {% ([int, frac, exp]) => int + (frac ? "." + (frac[1] ?? "") : "") + (exp ?? "") %}
        | "." digits exponent:?              {% ([, frac, exp]) => "." + frac + (exp ?? "") %}
digits -> [0-9]:+ {% ([d]) => d.join("") %}
exponent -> [eE] [+\-]:? digits {% ([e, sign, d]) => e + (sign ?? "") + d %}

_ -> [ \t\n\r]:* {% () => null %}
