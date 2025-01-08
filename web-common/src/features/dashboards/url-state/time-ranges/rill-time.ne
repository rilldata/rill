@preprocessor esmodule
@builtin "whitespace.ne"
@builtin "number.ne"
@builtin "string.ne"

@{%
  import {
    RillTimeAnchor,
    RillTime,
  } from "./RillTime.ts"
%}

rill_time => time_anchor_part                     {% ([{ start, end }]) => new RillTime(start, end) %}
           | time_anchor_part _ grain_and_at_part {% ([{ start, end }, , { grain, modifier }]) => new RillTime(start, end, grain, modifier) %}

time_anchor_part => time_anchor _ "," _ time_anchor {% ([start, , , , end]) => ({ start, end }) %}
                  | time_anchor                     {% ([start]) => ({ start }) %}

time_anchor => time_anchor_offset               {% id %}
             | time_anchor_offset _ "/" _ grain {% ([mod, , , , truncate]) => mod.withTruncate(truncate) %}

time_anchor_offset => "now"          {% () => RillTimeAnchor.now() %}
                    | "earliest"     {% () => RillTimeAnchor.earliest() %}
                    | "latest"       {% () => RillTimeAnchor.latest() %}
                    | grain_modifier {% ([grain]) => RillTimeAnchor.custom(grain) %}

grain_and_at_part => ":" _ range_grain_modifier _ "@" _ at_modifiers {% ([, , grain, , , , modifier]) => ({ grain, modifier }) %}
                   | ":" _ range_grain_modifier                      {% ([, , grain]) => ({ grain }) %}
                   | "@" _ at_modifiers                              {% ([, , modifier]) => ({ modifier }) %}

range_grain_modifier => grain             {% ([grain]) => ({ grain, isComplete: false }) %}
                      | "|" _ grain _ "|" {% ([, ,grain]) => ({ grain, isComplete: true }) %}

at_modifiers => grain_modifier                     {% ([grain]) => ({ at: RillTimeAnchor.custom(grain) }) %}
              | timezone_modifier                  {% ([timeZone]) => ({ timeZone }) %}
              | grain_modifier _ timezone_modifier {% ([grain, , timeZone]) => ({ at: RillTimeAnchor.custom(grain), timeZone }) %}

grain_modifier => grain     {% ([grain]) => ({ count: 0, grain }) %}
                | int grain {% ([count, grain]) => ({ count, grain }) %}

timezone_modifier => "{" _ [^}]:+ _ "}" {% ([, , tz]) => tz.join("") %}

grain => [smhdDWQMY] {% id %}
