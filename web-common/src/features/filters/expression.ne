# match an expression used in url filters
@builtin "whitespace.ne"
@builtin "number.ne"
@builtin "string.ne"

expr        => column __ "IN"i _ "(" value_list ")"      {% ([column, _1, op, _2, _3, values]) => [column, op, values] %}
             | column __ "NIN"i _ "(" value_list ")"     {% ([column, _1, op, _2, _3, values]) => [column, op, values] %}
             | column __ "HAVING"i _ "(" expr ")"        {% ([column, _1, op, _2, _3, expr]) => [column, op, expr] %}
             | expr_part _ binary_operator _ expr_value  {% ([left, _1, op, _2, right]) => [left, op, right] %}
             | expr_part _ compare_operator _ value      {% ([left, _1, op, _2, right]) => [left, op, right] %}

expr_part   => column       {% id %}
             | expr         {% id %}
             | "(" expr ")" {% ([_, expr]) => expr %}
expr_value  => value        {% id %}
             | expr         {% id %}
             | "(" expr ")" {% ([_, expr]) => expr %}

compare_operator => "="       {% id %}
                  | "!="      {% id %}
                  | ">"       {% id %}
                  | ">="      {% id %}
                  | "<"       {% id %}
                  | "<="      {% id %}
binary_operator  => "AND"i    {% id %}
                  | "OR"i     {% id %}

column      => sqstring                 {% id %}
            | [a-zA-Z] [a-zA-Z0-9_]:*   {% ([fst, rest]) => [fst, ...rest].join("") %}
value       => sqstring                 {% id %}
            | int                       {% id %}
            | decimal                   {% id %}
value_list  => value_list _ "," _ value {% ([list, _1, _2, _3, value]) => [...list, value] %}
            | value                     {% ([v]) => [v] %}
