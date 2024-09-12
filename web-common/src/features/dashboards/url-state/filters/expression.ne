# match an expression used in url filters
# adapted from https://dev.mysql.com/doc/refman/5.7/en/expressions.html
@builtin "whitespace.ne"
@builtin "number.ne"
@builtin "string.ne"

expr => expr _ "OR"i _ expr
      | expr _ "AND"i _ expr

bool_expr => 


expr => "(" expr ")"                               {% ([_, expr]) => expr %}
      | column __ in_operator _ "(" value_list ")" {% ([column, _1, op, _2, _3, values]) => [op.toUpperCase(), column, values] %}
      | column __ "HAVING"i _ "(" expr ")"         {% ([column, _1, op, _2, _3, expr]) => [op.toUpperCase(), column, expr] %}
      | expr _ "AND"i _ expr                       {% ([left, _1, op, _2, right]) => [op.toUpperCase(), left, right] %}
      | expr _ "OR"i _ expr                        {% ([left, _1, op, _2, right]) => [op.toUpperCase(), left, right] %}
      | or_expr                                    {% id %}
      | expr_or_col _ compare_operator _ value     {% ([left, _1, op, _2, right]) => [op.toUpperCase(), left, right] %}

expr_or_col => expr   {% id %}
             | column {% id %}

and_expr => expr _ "AND"i _ expr {% ([left, _1, op, _2, right]) => [op.toUpperCase(), left, right] %}
          # merge chained AND expressions
          | and_expr _ "AND"i _ expr {% ([left, _1, op, _2, right]) => [...left, right] %}

or_expr => expr _ "OR"i _ expr {% ([left, _1, op, _2, right]) => [op.toUpperCase(), left, right] %}
         # merge chained OR expressions
         | or_expr _ "OR"i _ expr {% ([left, _1, op, _2, right]) => [...left, right] %}

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
           | [a-zA-Z] [a-zA-Z0-9_]:*   {% ([fst, rest]) => [fst, ...rest].join("") %}
value      => sqstring                 {% id %}
           | int                       {% id %}
           | decimal                   {% id %}
value_list => value_list _ "," _ value {% ([list, _1, _2, _3, value]) => [...list, value] %}
           | value                     {% ([v]) => [v] %}
