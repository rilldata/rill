@preprocessor esmodule
@builtin "whitespace.ne"
@builtin "string.ne"

@{%
  import {
    RillTime,

    RillShorthandInterval,
    RillPeriodToGrainInterval,
    RillTimeStartEndInterval,
    RillTimeOrdinalInterval,
    RillIsoInterval,

    RillPointInTime,
    RillPointInTimeWithSnap,
    RillLabelledPointInTime,
    RillGrainPointInTime,
    RillGrainPointInTimePart,
    RillAbsoluteTime,
  } from "./RillTime.ts"
%}

rill_time => interval_with_grain                            {% id %}
           | interval_with_grain _ "tz" _ timezone_modifier {% ([rt, , , , tz]) => rt.withTimezone(tz) %}

interval_with_grain => interval_with_anchor_override _ "by"i _ grain {% ([rt, , , , grain]) => rt.withGrain(grain) %}
                     | interval_with_anchor_override                 {% id %}

interval_with_anchor_override => interval anchor_override:*      {% ([interval, anchorOverrides]) => new RillTime(interval).withAnchorOverrides(anchorOverrides) %}
anchor_override               => _ "as"i _ "of"i _ point_in_time {% ([, , , , , pointInTime]) => pointInTime %}

interval => shorthand_interval         {% id %}
          | period_to_grain_interval   {% id %}
          | start_end_interval         {% id %}
          | ordinal_interval           {% id %}
          | iso_interval               {% id %}

shorthand_interval => grain_duration {% ([parts]) => new RillShorthandInterval(parts) %}

period_to_grain_interval => period_to_grain {% ([grain]) => new RillPeriodToGrainInterval(grain) %}

ordinal_interval => ordinal (_ "of"i _ ordinal):* {% ([part, rest]) => new RillTimeOrdinalInterval([part, ...rest.map(([, , , p]) => p)]) %}

start_end_interval => point_in_time _ "to"i _ point_in_time {% ([start, , , , end]) => new RillTimeStartEndInterval(start, end) %}

iso_interval => abs_time _ "to"i _ abs_time {% ([start, , , , end]) => new RillIsoInterval(start, end) %}
              | abs_time _ "/" _ abs_time   {% ([start, , , , end]) => new RillIsoInterval(start, end) %}
              | abs_time _ "," _ abs_time   {% ([start, , , , end]) => new RillIsoInterval(start, end) %}
              | abs_time                    {% ([start]) => new RillIsoInterval(start, undefined) %}

point_in_time              => point_in_time_with_snap:* point_in_time_without_snap {% ([points, last]) => new RillPointInTime([...points, last]) %}
                            | point_in_time_with_snap                              {% ([point]) => new RillPointInTime([point]) %}
point_in_time_with_snap    => point_in_time_variants _ "/" _ grain _ "/" _ grain   {% ([point, , , , firstGrain, , , , secondGrain]) => new RillPointInTimeWithSnap(point, [firstGrain, secondGrain]) %}
                            | point_in_time_variants _ "/" _ grain                 {% ([point, , , , grain]) => new RillPointInTimeWithSnap(point, [grain]) %}
point_in_time_without_snap => point_in_time_variants                               {% ([point]) => new RillPointInTimeWithSnap(point, []) %}

point_in_time_variants => grain_point_in_time   {% id %}
                        | labeled_point_in_time {% id %}
                        | abs_time              {% id %}

grain_point_in_time      => grain_point_in_time_part:+ {% ([parts]) => new RillGrainPointInTime([...parts]) %}
grain_point_in_time_part => prefix _ grain_duration    {% ([prefix, _, grains]) => new RillGrainPointInTimePart(prefix, grains) %}

labeled_point_in_time => "earliest"i  {% RillLabelledPointInTime.postProcessor %}
                       | "latest"i    {% RillLabelledPointInTime.postProcessor %}
                       | "now"i       {% RillLabelledPointInTime.postProcessor %}
                       | "watermark"i {% RillLabelledPointInTime.postProcessor %}
                       | "ref"i       {% RillLabelledPointInTime.postProcessor %}

ordinal => grain num {% ([grain, num]) => ({num, grain}) %}

grain_duration      => grain_duration_part:+ {% ([parts]) => parts %}
grain_duration_part => num grain             {% ([num, grain]) => ({num, grain}) %}

period_to_grain => [sSmhHdDwWqQMyY] "TD" {% ([grain]) => grain %}

abs_time => [\d] [\d] [\d] [\d] [\-] [\d] [\d] [\-] [\d] [\d] "T" [\d] [\d] [:] [\d] [\d] [:] [\d] [\d] [.] [\d]:+ "Z" {% RillAbsoluteTime.postProcessor %}
          | [\d] [\d] [\d] [\d] [\-] [\d] [\d] [\-] [\d] [\d] "T" [\d] [\d] [:] [\d] [\d] [:] [\d] [\d] "Z"            {% RillAbsoluteTime.postProcessor %}
          | [\d] [\d] [\d] [\d] [\-] [\d] [\d] [\-] [\d] [\d] "T" [\d] [\d] [:] [\d] [\d]                              {% RillAbsoluteTime.postProcessor %}
          | [\d] [\d] [\d] [\d] [\-] [\d] [\d] [\-] [\d] [\d] "T" [\d] [\d]                                            {% RillAbsoluteTime.postProcessor %}
          | [\d] [\d] [\d] [\d] [\-] [\d] [\d] [\-] [\d] [\d]                                                          {% RillAbsoluteTime.postProcessor %}
          | [\d] [\d] [\d] [\d] [\-] [\d] [\d]                                                                         {% RillAbsoluteTime.postProcessor %}
          | [\d] [\d] [\d] [\d]                                                                                        {% RillAbsoluteTime.postProcessor %}

timezone_modifier => [0-9a-zA-Z/+\-_]:+ {% ([args]) => args.join("") %}

prefix => [+\-] {% id %}

num => [0-9]:+ {% ([args]) => Number(args.join("")) %}

grain => [sSmhHdDwWqQMyY] {% id %}
