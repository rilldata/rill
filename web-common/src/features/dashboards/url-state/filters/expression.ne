@preprocessor esmodule
# match an expression used in url filters
# adapted from https://dev.mysql.com/doc/refman/5.7/en/expressions.html
@builtin "whitespace.ne"
@builtin "number.ne"
@builtin "string.ne"

@{%
  import {
    binaryPostprocessor,
    inPostprocessor,
    havingPostprocessor,
    andOrPostprocessor,
  } from "./post-processors.ts";
%}

expr => boolean_expr                             {% id %}
      # chain of AND expressions with a terminal that doesnt immediately contain AND
      | (boolean_expr _ "AND"i _):+ non_and_expr {% andOrPostprocessor %}
      # chain of OR expressions with a terminal that doesnt immediately contain OR
      | (boolean_expr _ "OR"i _):+ non_or_expr   {% andOrPostprocessor %}

# these are used to disambiguate matches.
# these contain either boolean_expr or the other joiner (AND for non_and_expr, AND for non_or_expr)
non_and_expr => boolean_expr                              {% id %}
              | (boolean_expr __ "OR"i __):+ non_and_expr {% andOrPostprocessor %}
non_or_expr  => boolean_expr                              {% id %}
              | (boolean_expr __ "AND"i __):+ non_or_expr {% andOrPostprocessor %}

# expression that result in a boolean result excluding AND, OR
boolean_expr => "(" expr ")"                               {% ([_, expr]) => expr %}
              | column __ in_operator _ "(" value_list ")" {% inPostprocessor %}
              | column __ "HAVING"i _ "(" expr ")"         {% havingPostprocessor %}
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
