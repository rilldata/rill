# match an expression used in url filters
# adapted from https://dev.mysql.com/doc/refman/5.7/en/expressions.html
@builtin "whitespace.ne"
@builtin "number.ne"
@builtin "string.ne"

@{%
  const binaryPostprocessor = ([left, _1, op, _2, right]) => [op.toUpperCase(), left, right];
  // binary expression where the right should be surrounded in brackets
  const bracketedBinaryPostprocessor = ([column, _1, op, _2, _3, values]) => [op.toUpperCase(), column, values];
  const andOrPostprocessor = ([left, right]) => {
    const op = left[0][2].toUpperCase();
    const exprs = left.map((_, i) => i % 4 === 0);
    return [op, ...left.map((t) => t[0]), right]
  }
%}

expr => boolean_expr                             {% id %}
      | (boolean_expr _ "AND"i _):+ non_and_expr {% andOrPostprocessor %}
      | (boolean_expr _ "OR"i _):+ non_or_expr   {% andOrPostprocessor %}

# these are used to disambiguate matches
non_and_expr => boolean_expr                            {% id %}
              | (boolean_expr _ "OR"i _):+ non_and_expr {% andOrPostprocessor %}
non_or_expr  => boolean_expr                            {% id %}
              | (boolean_expr _ "AND"i _):+ non_or_expr {% andOrPostprocessor %}

boolean_expr => "(" expr ")"                               {% ([_, expr]) => expr %}
              | column __ in_operator _ "(" value_list ")" {% bracketedBinaryPostprocessor %}
              | column __ "HAVING"i _ "(" expr ")"         {% bracketedBinaryPostprocessor %}
              | simple_expr _ compare_operator _ value     {% binaryPostprocessor %}

simple_expr => column {% id %}
             | value  {% id %}

in_operator      => "IN"i     {% id %}
                  | "NIN"i    {% id %}
                  | "NOT IN"i {% () => "NIN" %}
compare_operator => "="       {% id %}
                  | "!="      {% id %}
                  | ">"       {% id %}
                  | ">="      {% id %}
                  | "<"       {% id %}
                  | "<="      {% id %}

column     => sqstring                 {% id %}
            | [a-zA-Z] [a-zA-Z0-9_]:*  {% ([fst, rest]) => [fst, ...rest].join("") %}
value      => sqstring                 {% id %}
            | int                      {% id %}
            | decimal                  {% id %}
value_list => value_list _ "," _ value {% ([list, _1, _2, _3, value]) => [...list, value] %}
            | value                    {% ([v]) => [v] %}
