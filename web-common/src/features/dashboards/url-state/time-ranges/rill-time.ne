@preprocessor esmodule
@builtin "whitespace.ne"
@builtin "string.ne"

@{%
  import {
    RillTime,

    RillTimeAnchoredDurationInterval,
    RillShorthandInterval,
    RillPeriodToGrainInterval,
    RillTimeStartEndInterval,
    RillTimeOrdinalInterval,
    RillGrainToInterval,
    RillIsoInterval,

    RillOrdinalPointInTime,
    RillOrdinalPart,
    RillOrdinalPartEnd,
    RillGrainPointInTime,
    RillGrainPointInTimePart,
    RillAbsoluteTime,
  } from "./RillTime.ts"
%}

rill_time => full_interval                            {% id %}
           | full_interval _ "tz" _ timezone_modifier {% ([rt, , , , tz]) => rt.withTimezone(tz) %}

full_interval => interval_with_grain _ "as"i _ "of"i _ anchor_override {% id %}
               | interval_with_grain                                   {% id %}

interval_with_grain => interval _ "by"i _ grain {% ([interval, , , , grain]) => new RillTime(interval).withGrain(grain) %}
                     | interval                 {% ([interval]) => new RillTime(interval) %}

interval => anchored_duration_interval {% id %}
          | shorthand_interval         {% id %}
          | period_to_grain_interval   {% id %}
          | start_end_interval         {% id %}
          | ordinal_interval           {% id %}
          | grain_to_interval          {% id %}
          | iso_interval               {% id %}

anchored_duration_interval => grain_duration _ "starting"i _ point_in_time {% ([grains, , , , point]) => new RillTimeAnchoredDurationInterval(grains, true, point) %}
                            | grain_duration _ "ending"i _ point_in_time   {% ([grains, , , , point]) => new RillTimeAnchoredDurationInterval(grains, false, point) %}

shorthand_interval => num grain _ "in"i _ grain "!" {% ([num, grain, , , , inGrain]) => new RillShorthandInterval(num, grain, inGrain, false) %}
                    | num grain _ "in"i _ grain     {% ([num, grain, , , , inGrain]) => new RillShorthandInterval(num, grain, inGrain, true) %}
                    | num grain "!"                 {% ([num, grain]) => new RillShorthandInterval(num, grain, undefined, false) %}
                    | num grain                     {% ([num, grain]) => new RillShorthandInterval(num, grain, undefined, true) %}

period_to_grain_interval => period_to_grain _ "in"i _ grain "!" {% ([grain, , , , inGrain]) => new RillPeriodToGrainInterval(grain, inGrain, false) %}
                          | period_to_grain _ "in"i _ grain     {% ([grain, , , , inGrain]) => new RillPeriodToGrainInterval(grain, inGrain, true) %}
                          | period_to_grain "!"                 {% ([grain]) => new RillPeriodToGrainInterval(grain, undefined, false) %}
                          | period_to_grain                     {% ([grain]) => new RillPeriodToGrainInterval(grain, undefined, true) %}

ordinal_interval => ordinal_duration _ "of"i _ ordinal_interval_end {% ([parts, , , , end]) => new RillTimeOrdinalInterval(parts, end) %}
                  | ordinal_duration                                {% ([parts, , , , end]) => new RillTimeOrdinalInterval(parts, undefined) %}

ordinal_interval_end => grain_to_interval  {% ([grainToInterval]) => new RillOrdinalPartEnd().withGrainToInterval(grainToInterval) %}
                      | start_end_interval {% ([startEndInterval]) => new RillOrdinalPartEnd().withStartEndInterval(startEndInterval) %}
                      | grain              {% ([grain]) => new RillOrdinalPartEnd().withSingleGrain(grain) %}

start_end_interval => point_in_time _ "to"i _ point_in_time {% ([start, , , , end]) => new RillTimeStartEndInterval(start, end) %}

grain_to_interval => grain_point_in_time _ "#" {% ([point]) => new RillGrainToInterval(point) %}

iso_interval => abs_time _ "to"i _ abs_time {% ([start, , , , end]) => new RillIsoInterval(start, end) %}
              | abs_time _ "/" _ abs_time   {% ([start, , , , end]) => new RillIsoInterval(start, end) %}
              | abs_time                    {% ([start]) => new RillIsoInterval(start, undefined) %}

anchor_override => grain_point_in_time
                 | labeled_point_in_time
                 | abs_time

point_in_time => ordinal_point_in_time {% id %}
               | grain_point_in_time   {% id %}
               | labeled_point_in_time {% id %}

ordinal_point_in_time     => ordinal_point_in_time_start _ "of"i _ ordinal_duration _ "of"i _ ordinal_point_in_time_end {% ([start, , , , parts, , , ,end]) => start.withRestOfParts(parts).withEnd(end) %}
                           | ordinal_point_in_time_start _ "of"i _ ordinal_duration                                     {% ([start, , , , parts]) => start.withRestOfParts(parts) %}
                           | ordinal_point_in_time_start _ "of"i _ ordinal_point_in_time_end                            {% ([start, , , , end]) => start.withEnd(end) %}
                           | ordinal_point_in_time_start                                                                {% id %}

ordinal_point_in_time_start => ordinal_part _ suffix {% ([ordinal, suffix]) => new RillOrdinalPointInTime(ordinal, suffix) %}

ordinal_point_in_time_end => grain_to_interval {% ([grainToInterval]) => new RillOrdinalPartEnd().withGrainToInterval(grainToInterval) %}
                           | grain             {% ([grain]) => new RillOrdinalPartEnd().withSingleGrain(grain) %}

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

ordinal_duration => ordinal_part (_ "of"i _ ordinal_part):* {% ([part, rest]) => ([part, ...rest.map(([, , , p]) => p)]) %}

ordinal_part => ordinal                            {% ([{num, grain}]) => new RillOrdinalPart(grain, num, undefined) %}
              | snap_prefix _ grain_duration_part  {% ([snap, , {num, grain}]) => new RillOrdinalPart(grain, num, snap) %}
ordinal      => grain num                          {% ([grain, num]) => ({num, grain}) %}

grain_duration      => grain_duration_part (_ grain_duration_part):* {% ([part, rest]) => ([part, ...rest.map(([, p]) => p)]) %}
grain_duration_part => num grain                                     {% ([num, grain]) => ({num, grain}) %}
                     | grain                                         {% ([grain]) => ({grain}) %}

period_to_grain => [sSmhHdDwWqQMyY] "TD" {% ([grain]) => grain %}

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

