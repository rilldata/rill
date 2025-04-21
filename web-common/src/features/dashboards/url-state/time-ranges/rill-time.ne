@preprocessor esmodule
@builtin "whitespace.ne"
@builtin "string.ne"

@{%
  import {
    RillTime,
    RillTimeAbsoluteTime,
    RillTimeLabelledAnchor,
    RillTimeOrdinal,
    RillTimeRelative,
    RillTimePeriodToDate,
  } from "./RillTime.ts"
%}

rill_time => time_range                       {% id %}
           | time_range "@" timezone_modifier {% id %}

time_range => link _ "to"i _ link _ "by"i _ grain {% id %}
            | link _ "to"i _ link                 {% ([start, , , , end]) => new RillTime(start, end, undefined) %}
            | link _ "by"i _ grain                {% id %}
            | link                                {% ([start]) => new RillTime(start, undefined, undefined) %}

link => link_part non_first_link_part:* {% ([first, rest]) => [first, ...rest] %}

non_first_link_part => _ "of"i _ link_part {% ([, , , lp]) => lp %}

link_part => labeled_anchor             {% id %}
           | abs_time                   {% id %}
           | ordinal                    {% id %}
           | period_to_grain_with_tilde {% id %}
           | relative_time_with_tilde   {% id %}

ordinal => grain num {% ([grain, num]) => new RillTimeOrdinal(grain, num) %}

relative_time_with_tilde => relative_time     {% id %}
                          | relative_time "~" {% ([rt]) => rt.asIncomplete() %}

relative_time => grain            {% ([grain]) => new RillTimeRelative(undefined, 1, grain) %}
               | prefix grain     {% ([prefix, grain]) => new RillTimeRelative(prefix, 1, grain) %}
               | num grain        {% ([num, grain]) => new RillTimeRelative(undefined, num, grain) %}
               | prefix num grain {% ([prefix, num, grain]) => new RillTimeRelative(prefix, num, grain) %}

period_to_grain_with_tilde => period_to_grain_time     {% id %}
                            | period_to_grain_time "~" {% ([ptd]) => ptd.asIncomplete() %}

period_to_grain_time => period_to_grain            {% ([ptg]) => new RillTimePeriodToDate(undefined, 1, ptg) %}
                      | prefix period_to_grain     {% ([prefix, ptg]) => new RillTimePeriodToDate(prefix, 1, ptg) %}
                      | num period_to_grain        {% ([num, ptg]) => new RillTimePeriodToDate(undefined, num, ptg) %}
                      | prefix num period_to_grain {% ([prefix, num, ptg]) => new RillTimePeriodToDate(prefix, num, ptg) %}

labeled_anchor => "earliest"  {% RillTimeLabelledAnchor.postProcessor %}
                | "latest"    {% RillTimeLabelledAnchor.postProcessor %}
                | "now"       {% RillTimeLabelledAnchor.postProcessor %}
                | "watermark" {% RillTimeLabelledAnchor.postProcessor %}

period_to_grain => [sSmhHdDwWqQMyY] [T] [sSmhHdDwWqQMyY] {% (args) => args.join("") %}

abs_time => [\d] [\d] [\d] [\d] [\-] [\d] [\d] [\-] [\d] [\d] "T" [\d] [\d] [:] [\d] [\d] [:] [\d] [\d] "Z" {% RillTimeAbsoluteTime.postProcessor %}
          | [\d] [\d] [\d] [\d] [\-] [\d] [\d] [\-] [\d] [\d] "T" [\d] [\d] [:] [\d] [\d]                   {% RillTimeAbsoluteTime.postProcessor %}
          | [\d] [\d] [\d] [\d] [\-] [\d] [\d] [\-] [\d] [\d] "T" [\d] [\d]                                 {% RillTimeAbsoluteTime.postProcessor %}
          | [\d] [\d] [\d] [\d] [\-] [\d] [\d] [\-] [\d] [\d]                                               {% RillTimeAbsoluteTime.postProcessor %}
          | [\d] [\d] [\d] [\d] [\-] [\d] [\d]                                                              {% RillTimeAbsoluteTime.postProcessor %}
          | [\d] [\d] [\d] [\d]                                                                             {% RillTimeAbsoluteTime.postProcessor %}

timezone_modifier => "{" _ [^}]:+ _ "}" {% ([, , tz]) => tz.join("") %}

prefix => [+\-<>] {% id %}

num => [0-9]:+ {% ([args]) => Number(args.join("")) %}

grain => [sSmhHdDwWqQMyY] {% id %}
