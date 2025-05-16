@preprocessor esmodule
@builtin "whitespace.ne"
@builtin "string.ne"

@{%
  import {
    RillTime,

    RillTimeAnchoredDurationInterval,
    RillTimeStartEndInterval,
    RillGrainToInterval,

    RillGrainPointInTime,
    RillGrainPointInTimePart,
    RillAbsoluteTime,
  } from "./RillTime.ts"
%}

rill_time => full_interval                            {% id %}
           | full_interval _ "tz" _ timezone_modifier {% id %}

full_interval => interval_with_grain _ "as"i _ "of"i _ anchor_override {% id %}
               | interval_with_grain                                   {% id %}

interval_with_grain => interval _ "by"i _ grain {% ([interval, , , , grain]) => new RillTime(interval).withGrain(grain) %}
                     | interval                 {% ([interval]) => new RillTime(interval) %}

interval => anchored_duration_interval {% id %}
          | ordinal_interval           {% id %}
          | start_end_interval         {% id %}
          | grain_to_interval          {% id %}
          | iso_interval               {% id %}

anchored_duration_interval => grain_duration _ "starting"i _ point_in_time {% ([grains, , , , point]) => new RillTimeAnchoredDurationInterval(grains, true, point) %}
                            | grain_duration _ "ending"i _ point_in_time   {% ([grains, , , , point]) => new RillTimeAnchoredDurationInterval(grains, false, point) %}

ordinal_interval => ordinal_duration _ "of"i _ ordinal_interval_end
                  | ordinal_duration
ordinal_interval_end => grain_to_interval
                      | start_end_interval
                      | grain

start_end_interval => point_in_time _ "to"i _ point_in_time {% ([start, , , , end]) => new RillTimeStartEndInterval(start, end) %}

grain_to_interval => grain_point_in_time _ "!" {% ([point]) => new RillGrainToInterval(point) %}

iso_interval => abs_time _ "to"i _ abs_time
              | abs_time _ "/" _ abs_time
              | abs_time

anchor_override => grain_point_in_time
                 | labeled_point_in_time
                 | abs_time

point_in_time => ordinal_point_in_time {% id %}
               | grain_point_in_time   {% id %}
               | labeled_point_in_time {% id %}

ordinal_point_in_time => ordinal _ suffix _ ordinal_duration
                       | ordinal _ suffix

grain_point_in_time => grain_point_in_time_part (_ prefixed_grain_point_in_time_part):* {% ([part, rest]) => new RillGrainPointInTime([part, ...rest.map(([, p]) => p)]) %}

grain_point_in_time_part            => prefixed_grain_point_in_time_part            {% id %}
                                     | unprefixed_grain_point_in_time_part          {% id %}
prefixed_grain_point_in_time_part   => prefix _ unprefixed_grain_point_in_time_part {% ([prefix, _, point]) => point.withPrefix(prefix) %}
unprefixed_grain_point_in_time_part => grain_duration _ "/" _ snap_grain _ suffix   {% ([grains, , , , snap, , suffix]) => new RillGrainPointInTimePart(grains).withSnap(snap).withSuffix(suffix) %}
                                     | grain_duration _ "/" _ snap_grain            {% ([grains, , , , snap]) => new RillGrainPointInTimePart(grains).withSnap(snap) %}
                                     | grain_duration _ suffix                      {% ([grains, , suffix]) => new RillGrainPointInTimePart(grains).withSuffix(suffix) %}
                                     | grain_duration                               {% ([grains]) => new RillGrainPointInTimePart(grains) %}

labeled_point_in_time => "earliest"  {% id %}
                       | "latest"    {% id %}
                       | "now"       {% id %}
                       | "watermark" {% id %}

ordinal_duration      => ordinal_duration_part (_ "of"i _ ordinal_duration_part):*
ordinal_duration_part => ordinal
                       | snap_prefix _ grain_duration_part

ordinal => grain num

grain_duration      => grain_duration_part (_ grain_duration_part):* {% ([part, rest]) => ([part, ...rest.map(([, p]) => p)]) %}
grain_duration_part => num grain                                     {% ([num, grain]) => ({num, grain}) %}
                     | grain                                         {% ([grain]) => ({grain}) %}

abs_time => [\d] [\d] [\d] [\d] [\-] [\d] [\d] [\-] [\d] [\d] "T" [\d] [\d] [:] [\d] [\d] [:] [\d] [\d] "Z" {% RillAbsoluteTime.postProcessor %}
          | [\d] [\d] [\d] [\d] [\-] [\d] [\d] [\-] [\d] [\d] "T" [\d] [\d] [:] [\d] [\d]                   {% RillAbsoluteTime.postProcessor %}
          | [\d] [\d] [\d] [\d] [\-] [\d] [\d] [\-] [\d] [\d] "T" [\d] [\d]                                 {% RillAbsoluteTime.postProcessor %}
          | [\d] [\d] [\d] [\d] [\-] [\d] [\d] [\-] [\d] [\d]                                               {% RillAbsoluteTime.postProcessor %}
          | [\d] [\d] [\d] [\d] [\-] [\d] [\d]                                                              {% RillAbsoluteTime.postProcessor %}
          | [\d] [\d] [\d] [\d]                                                                             {% RillAbsoluteTime.postProcessor %}

timezone_modifier => [0-9a-zA-Z/+\-_]:+ {% ([args]) => args.join("") %}

prefix => [+\-] {% id %}

snap_prefix => [<>] {% id %}

suffix => [\^\$] {% id %}

num => [0-9]:+ {% ([args]) => Number(args.join("")) %}

snap_grain      => week_snap_grain {% id %}
                 | grain           {% id %}
week_snap_grain => [qQMyY] [wW]    {% ([args]) => args.join("") %}

grain => [sSmhHdDwWqQMyY] {% id %}

