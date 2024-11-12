@preprocessor esmodule
@builtin "whitespace.ne"
@builtin "number.ne"
@builtin "string.ne"

@{%
  import {
    RillTimeModifier,
    RillTime,
  } from "./RillTime.ts"
%}

rill_time => time_mod _ "," _ time_mod _ ":" _ modifiers {% ([start, , , , end, , , , modifiers]) => new RillTime(start, end, modifiers) %}
           | time_mod _ "," _ time_mod                   {% ([start, , , , end]) => new RillTime(start, end) %}
           | time_mod _ ":" _ modifiers                  {% ([start, , , , modifiers]) => new RillTime(start, undefined, modifiers) %}
           | time_mod                                    {% ([start]) => new RillTime(start) %}

time_mod => time_mod_offset               {% id %}
          | time_mod_offset _ "/" _ grain {% ([mod, , , , truncate]) => mod.withTruncate(truncate) %}

time_mod_offset => "now"          {% () => RillTimeModifier.now() %}
                 | "earliest"     {% () => RillTimeModifier.earliest() %}
                 | "latest"       {% () => RillTimeModifier.latest() %}
                 | grain_modifier {% ([grain]) => RillTimeModifier.custom(grain) %}

modifiers => range_grain_modifier                {% ([timeRangeGrain]) => ({ timeRangeGrain }) %}
           | range_grain_modifier _ at_modifiers {% ([timeRangeGrain, , atModifiers]) => ({ timeRangeGrain, ...atModifiers }) %}
           | at_modifiers                        {% id %}

at_modifiers => "@" _ grain_modifier                     {% ([, , grain]) => ({ grain }) %}
              | "@" _ timezone_modifier                  {% ([, , timeZone]) => ({ timeZone }) %}
              | "@" _ grain_modifier _ timezone_modifier {% ([, , timeZone]) => ({ timeZone }) %}

grain_modifier => grain     {% ([grain]) => ({ count: 0, grain }) %}
                | int grain {% ([count, grain]) => ({ count, grain }) %}

timezone_modifier => "{" _ [a-zA-Z]:+ _ "}" {% ([, , tz]) => tz %}

range_grain_modifier => grain             {% ([grain]) => ({ grain, isComplete: false }) %}
                      | "|" _ grain _ "|" {% ([, ,grain]) => ({ grain, isComplete: true }) %}

grain => [smhdDWQMY] {% id %}
