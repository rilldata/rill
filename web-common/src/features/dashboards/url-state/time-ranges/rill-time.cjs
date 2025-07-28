// Generated automatically by nearley, version 2.20.1
// http://github.com/Hardmath123/nearley
// eslint-disable
function id(x) {
  return x[0];
}

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
} from "./RillTime.ts";
let Lexer = undefined;
let ParserRules = [
  { name: "_$ebnf$1", symbols: [] },
  {
    name: "_$ebnf$1",
    symbols: ["_$ebnf$1", "wschar"],
    postprocess: function arrpush(d) {
      return d[0].concat([d[1]]);
    },
  },
  {
    name: "_",
    symbols: ["_$ebnf$1"],
    postprocess: function (d) {
      return null;
    },
  },
  { name: "__$ebnf$1", symbols: ["wschar"] },
  {
    name: "__$ebnf$1",
    symbols: ["__$ebnf$1", "wschar"],
    postprocess: function arrpush(d) {
      return d[0].concat([d[1]]);
    },
  },
  {
    name: "__",
    symbols: ["__$ebnf$1"],
    postprocess: function (d) {
      return null;
    },
  },
  { name: "wschar", symbols: [/[ \t\n\v\f]/], postprocess: id },
  { name: "dqstring$ebnf$1", symbols: [] },
  {
    name: "dqstring$ebnf$1",
    symbols: ["dqstring$ebnf$1", "dstrchar"],
    postprocess: function arrpush(d) {
      return d[0].concat([d[1]]);
    },
  },
  {
    name: "dqstring",
    symbols: [{ literal: '"' }, "dqstring$ebnf$1", { literal: '"' }],
    postprocess: function (d) {
      return d[1].join("");
    },
  },
  { name: "sqstring$ebnf$1", symbols: [] },
  {
    name: "sqstring$ebnf$1",
    symbols: ["sqstring$ebnf$1", "sstrchar"],
    postprocess: function arrpush(d) {
      return d[0].concat([d[1]]);
    },
  },
  {
    name: "sqstring",
    symbols: [{ literal: "'" }, "sqstring$ebnf$1", { literal: "'" }],
    postprocess: function (d) {
      return d[1].join("");
    },
  },
  { name: "btstring$ebnf$1", symbols: [] },
  {
    name: "btstring$ebnf$1",
    symbols: ["btstring$ebnf$1", /[^`]/],
    postprocess: function arrpush(d) {
      return d[0].concat([d[1]]);
    },
  },
  {
    name: "btstring",
    symbols: [{ literal: "`" }, "btstring$ebnf$1", { literal: "`" }],
    postprocess: function (d) {
      return d[1].join("");
    },
  },
  { name: "dstrchar", symbols: [/[^\\"\n]/], postprocess: id },
  {
    name: "dstrchar",
    symbols: [{ literal: "\\" }, "strescape"],
    postprocess: function (d) {
      return JSON.parse('"' + d.join("") + '"');
    },
  },
  { name: "sstrchar", symbols: [/[^\\'\n]/], postprocess: id },
  {
    name: "sstrchar",
    symbols: [{ literal: "\\" }, "strescape"],
    postprocess: function (d) {
      return JSON.parse('"' + d.join("") + '"');
    },
  },
  {
    name: "sstrchar$string$1",
    symbols: [{ literal: "\\" }, { literal: "'" }],
    postprocess: function joiner(d) {
      return d.join("");
    },
  },
  {
    name: "sstrchar",
    symbols: ["sstrchar$string$1"],
    postprocess: function (d) {
      return "'";
    },
  },
  { name: "strescape", symbols: [/["\\/bfnrt]/], postprocess: id },
  {
    name: "strescape",
    symbols: [
      { literal: "u" },
      /[a-fA-F0-9]/,
      /[a-fA-F0-9]/,
      /[a-fA-F0-9]/,
      /[a-fA-F0-9]/,
    ],
    postprocess: function (d) {
      return d.join("");
    },
  },
  { name: "rill_time", symbols: ["interval_with_grain"], postprocess: id },
  {
    name: "rill_time$string$1",
    symbols: [{ literal: "t" }, { literal: "z" }],
    postprocess: function joiner(d) {
      return d.join("");
    },
  },
  {
    name: "rill_time",
    symbols: [
      "interval_with_grain",
      "_",
      "rill_time$string$1",
      "_",
      "timezone_modifier",
    ],
    postprocess: ([rt, , , , tz]) => rt.withTimezone(tz),
  },
  {
    name: "interval_with_grain$subexpression$1",
    symbols: [/[bB]/, /[yY]/],
    postprocess: function (d) {
      return d.join("");
    },
  },
  {
    name: "interval_with_grain",
    symbols: [
      "interval_with_anchor_override",
      "_",
      "interval_with_grain$subexpression$1",
      "_",
      "grain",
    ],
    postprocess: ([rt, , , , grain]) => rt.withGrain(grain),
  },
  {
    name: "interval_with_grain",
    symbols: ["interval_with_anchor_override"],
    postprocess: id,
  },
  { name: "interval_with_anchor_override$ebnf$1", symbols: [] },
  {
    name: "interval_with_anchor_override$ebnf$1",
    symbols: ["interval_with_anchor_override$ebnf$1", "anchor_override"],
    postprocess: function arrpush(d) {
      return d[0].concat([d[1]]);
    },
  },
  {
    name: "interval_with_anchor_override",
    symbols: ["interval", "interval_with_anchor_override$ebnf$1"],
    postprocess: ([interval, anchorOverrides]) =>
      new RillTime(interval).withAnchorOverrides(anchorOverrides),
  },
  {
    name: "anchor_override$subexpression$1",
    symbols: [/[aA]/, /[sS]/],
    postprocess: function (d) {
      return d.join("");
    },
  },
  {
    name: "anchor_override$subexpression$2",
    symbols: [/[oO]/, /[fF]/],
    postprocess: function (d) {
      return d.join("");
    },
  },
  {
    name: "anchor_override",
    symbols: [
      "_",
      "anchor_override$subexpression$1",
      "_",
      "anchor_override$subexpression$2",
      "_",
      "point_in_time",
    ],
    postprocess: ([, , , , , pointInTime]) => pointInTime,
  },
  { name: "interval", symbols: ["shorthand_interval"], postprocess: id },
  { name: "interval", symbols: ["period_to_grain_interval"], postprocess: id },
  { name: "interval", symbols: ["start_end_interval"], postprocess: id },
  { name: "interval", symbols: ["ordinal_interval"], postprocess: id },
  { name: "interval", symbols: ["iso_interval"], postprocess: id },
  {
    name: "shorthand_interval",
    symbols: ["grain_duration"],
    postprocess: ([parts]) => new RillShorthandInterval(parts),
  },
  {
    name: "period_to_grain_interval",
    symbols: ["period_to_grain"],
    postprocess: ([grain]) => new RillPeriodToGrainInterval(grain),
  },
  { name: "ordinal_interval$ebnf$1", symbols: [] },
  {
    name: "ordinal_interval$ebnf$1$subexpression$1$subexpression$1",
    symbols: [/[oO]/, /[fF]/],
    postprocess: function (d) {
      return d.join("");
    },
  },
  {
    name: "ordinal_interval$ebnf$1$subexpression$1",
    symbols: [
      "_",
      "ordinal_interval$ebnf$1$subexpression$1$subexpression$1",
      "_",
      "ordinal",
    ],
  },
  {
    name: "ordinal_interval$ebnf$1",
    symbols: [
      "ordinal_interval$ebnf$1",
      "ordinal_interval$ebnf$1$subexpression$1",
    ],
    postprocess: function arrpush(d) {
      return d[0].concat([d[1]]);
    },
  },
  {
    name: "ordinal_interval",
    symbols: ["ordinal", "ordinal_interval$ebnf$1"],
    postprocess: ([part, rest]) =>
      new RillTimeOrdinalInterval([part, ...rest.map(([, , , p]) => p)]),
  },
  {
    name: "start_end_interval$subexpression$1",
    symbols: [/[tT]/, /[oO]/],
    postprocess: function (d) {
      return d.join("");
    },
  },
  {
    name: "start_end_interval",
    symbols: [
      "point_in_time",
      "_",
      "start_end_interval$subexpression$1",
      "_",
      "point_in_time",
    ],
    postprocess: ([start, , , , end]) =>
      new RillTimeStartEndInterval(start, end),
  },
  {
    name: "iso_interval$subexpression$1",
    symbols: [/[tT]/, /[oO]/],
    postprocess: function (d) {
      return d.join("");
    },
  },
  {
    name: "iso_interval",
    symbols: ["abs_time", "_", "iso_interval$subexpression$1", "_", "abs_time"],
    postprocess: ([start, , , , end]) => new RillIsoInterval(start, end),
  },
  {
    name: "iso_interval",
    symbols: ["abs_time", "_", { literal: "/" }, "_", "abs_time"],
    postprocess: ([start, , , , end]) => new RillIsoInterval(start, end),
  },
  {
    name: "iso_interval",
    symbols: ["abs_time"],
    postprocess: ([start]) => new RillIsoInterval(start, undefined),
  },
  { name: "point_in_time$ebnf$1", symbols: [] },
  {
    name: "point_in_time$ebnf$1",
    symbols: ["point_in_time$ebnf$1", "point_in_time_with_snap"],
    postprocess: function arrpush(d) {
      return d[0].concat([d[1]]);
    },
  },
  {
    name: "point_in_time",
    symbols: ["point_in_time$ebnf$1", "point_in_time_without_snap"],
    postprocess: ([points, last]) => new RillPointInTime([...points, last]),
  },
  {
    name: "point_in_time",
    symbols: ["point_in_time_with_snap"],
    postprocess: ([point]) => new RillPointInTime([point]),
  },
  {
    name: "point_in_time_with_snap",
    symbols: [
      "point_in_time_variants",
      "_",
      { literal: "/" },
      "_",
      "grain",
      "_",
      { literal: "/" },
      "_",
      "grain",
    ],
    postprocess: ([point, , , , firstGrain, , , , secondGrain]) =>
      new RillPointInTimeWithSnap(point, [firstGrain, secondGrain]),
  },
  {
    name: "point_in_time_with_snap",
    symbols: ["point_in_time_variants", "_", { literal: "/" }, "_", "grain"],
    postprocess: ([point, , , , grain]) =>
      new RillPointInTimeWithSnap(point, [grain]),
  },
  {
    name: "point_in_time_without_snap",
    symbols: ["point_in_time_variants"],
    postprocess: ([point]) => new RillPointInTimeWithSnap(point, []),
  },
  {
    name: "point_in_time_variants",
    symbols: ["grain_point_in_time"],
    postprocess: id,
  },
  {
    name: "point_in_time_variants",
    symbols: ["labeled_point_in_time"],
    postprocess: id,
  },
  { name: "point_in_time_variants", symbols: ["abs_time"], postprocess: id },
  { name: "grain_point_in_time$ebnf$1", symbols: ["grain_point_in_time_part"] },
  {
    name: "grain_point_in_time$ebnf$1",
    symbols: ["grain_point_in_time$ebnf$1", "grain_point_in_time_part"],
    postprocess: function arrpush(d) {
      return d[0].concat([d[1]]);
    },
  },
  {
    name: "grain_point_in_time",
    symbols: ["grain_point_in_time$ebnf$1"],
    postprocess: ([parts]) => new RillGrainPointInTime([...parts]),
  },
  {
    name: "grain_point_in_time_part",
    symbols: ["prefix", "_", "grain_duration"],
    postprocess: ([prefix, _, grains]) =>
      new RillGrainPointInTimePart(prefix, grains),
  },
  {
    name: "labeled_point_in_time$subexpression$1",
    symbols: [/[eE]/, /[aA]/, /[rR]/, /[lL]/, /[iI]/, /[eE]/, /[sS]/, /[tT]/],
    postprocess: function (d) {
      return d.join("");
    },
  },
  {
    name: "labeled_point_in_time",
    symbols: ["labeled_point_in_time$subexpression$1"],
    postprocess: RillLabelledPointInTime.postProcessor,
  },
  {
    name: "labeled_point_in_time$subexpression$2",
    symbols: [/[lL]/, /[aA]/, /[tT]/, /[eE]/, /[sS]/, /[tT]/],
    postprocess: function (d) {
      return d.join("");
    },
  },
  {
    name: "labeled_point_in_time",
    symbols: ["labeled_point_in_time$subexpression$2"],
    postprocess: RillLabelledPointInTime.postProcessor,
  },
  {
    name: "labeled_point_in_time$subexpression$3",
    symbols: [/[nN]/, /[oO]/, /[wW]/],
    postprocess: function (d) {
      return d.join("");
    },
  },
  {
    name: "labeled_point_in_time",
    symbols: ["labeled_point_in_time$subexpression$3"],
    postprocess: RillLabelledPointInTime.postProcessor,
  },
  {
    name: "labeled_point_in_time$subexpression$4",
    symbols: [
      /[wW]/,
      /[aA]/,
      /[tT]/,
      /[eE]/,
      /[rR]/,
      /[mM]/,
      /[aA]/,
      /[rR]/,
      /[kK]/,
    ],
    postprocess: function (d) {
      return d.join("");
    },
  },
  {
    name: "labeled_point_in_time",
    symbols: ["labeled_point_in_time$subexpression$4"],
    postprocess: RillLabelledPointInTime.postProcessor,
  },
  {
    name: "labeled_point_in_time$subexpression$5",
    symbols: [/[rR]/, /[eE]/, /[fF]/],
    postprocess: function (d) {
      return d.join("");
    },
  },
  {
    name: "labeled_point_in_time",
    symbols: ["labeled_point_in_time$subexpression$5"],
    postprocess: RillLabelledPointInTime.postProcessor,
  },
  {
    name: "ordinal",
    symbols: ["grain", "num"],
    postprocess: ([grain, num]) => ({ num, grain }),
  },
  { name: "grain_duration$ebnf$1", symbols: ["grain_duration_part"] },
  {
    name: "grain_duration$ebnf$1",
    symbols: ["grain_duration$ebnf$1", "grain_duration_part"],
    postprocess: function arrpush(d) {
      return d[0].concat([d[1]]);
    },
  },
  {
    name: "grain_duration",
    symbols: ["grain_duration$ebnf$1"],
    postprocess: ([parts]) => parts,
  },
  {
    name: "grain_duration_part",
    symbols: ["num", "grain"],
    postprocess: ([num, grain]) => ({ num, grain }),
  },
  {
    name: "period_to_grain$string$1",
    symbols: [{ literal: "T" }, { literal: "D" }],
    postprocess: function joiner(d) {
      return d.join("");
    },
  },
  {
    name: "period_to_grain",
    symbols: [/[sSmhHdDwWqQMyY]/, "period_to_grain$string$1"],
    postprocess: ([grain]) => grain,
  },
  { name: "abs_time$ebnf$1", symbols: [/[\d]/] },
  {
    name: "abs_time$ebnf$1",
    symbols: ["abs_time$ebnf$1", /[\d]/],
    postprocess: function arrpush(d) {
      return d[0].concat([d[1]]);
    },
  },
  {
    name: "abs_time",
    symbols: [
      /[\d]/,
      /[\d]/,
      /[\d]/,
      /[\d]/,
      /[\-]/,
      /[\d]/,
      /[\d]/,
      /[\-]/,
      /[\d]/,
      /[\d]/,
      { literal: "T" },
      /[\d]/,
      /[\d]/,
      /[:]/,
      /[\d]/,
      /[\d]/,
      /[:]/,
      /[\d]/,
      /[\d]/,
      /[.]/,
      "abs_time$ebnf$1",
      { literal: "Z" },
    ],
    postprocess: RillAbsoluteTime.postProcessor,
  },
  {
    name: "abs_time",
    symbols: [
      /[\d]/,
      /[\d]/,
      /[\d]/,
      /[\d]/,
      /[\-]/,
      /[\d]/,
      /[\d]/,
      /[\-]/,
      /[\d]/,
      /[\d]/,
      { literal: "T" },
      /[\d]/,
      /[\d]/,
      /[:]/,
      /[\d]/,
      /[\d]/,
      /[:]/,
      /[\d]/,
      /[\d]/,
      { literal: "Z" },
    ],
    postprocess: RillAbsoluteTime.postProcessor,
  },
  {
    name: "abs_time",
    symbols: [
      /[\d]/,
      /[\d]/,
      /[\d]/,
      /[\d]/,
      /[\-]/,
      /[\d]/,
      /[\d]/,
      /[\-]/,
      /[\d]/,
      /[\d]/,
      { literal: "T" },
      /[\d]/,
      /[\d]/,
      /[:]/,
      /[\d]/,
      /[\d]/,
    ],
    postprocess: RillAbsoluteTime.postProcessor,
  },
  {
    name: "abs_time",
    symbols: [
      /[\d]/,
      /[\d]/,
      /[\d]/,
      /[\d]/,
      /[\-]/,
      /[\d]/,
      /[\d]/,
      /[\-]/,
      /[\d]/,
      /[\d]/,
      { literal: "T" },
      /[\d]/,
      /[\d]/,
    ],
    postprocess: RillAbsoluteTime.postProcessor,
  },
  {
    name: "abs_time",
    symbols: [
      /[\d]/,
      /[\d]/,
      /[\d]/,
      /[\d]/,
      /[\-]/,
      /[\d]/,
      /[\d]/,
      /[\-]/,
      /[\d]/,
      /[\d]/,
    ],
    postprocess: RillAbsoluteTime.postProcessor,
  },
  {
    name: "abs_time",
    symbols: [/[\d]/, /[\d]/, /[\d]/, /[\d]/, /[\-]/, /[\d]/, /[\d]/],
    postprocess: RillAbsoluteTime.postProcessor,
  },
  {
    name: "abs_time",
    symbols: [/[\d]/, /[\d]/, /[\d]/, /[\d]/],
    postprocess: RillAbsoluteTime.postProcessor,
  },
  { name: "timezone_modifier$ebnf$1", symbols: [/[0-9a-zA-Z/+\-_]/] },
  {
    name: "timezone_modifier$ebnf$1",
    symbols: ["timezone_modifier$ebnf$1", /[0-9a-zA-Z/+\-_]/],
    postprocess: function arrpush(d) {
      return d[0].concat([d[1]]);
    },
  },
  {
    name: "timezone_modifier",
    symbols: ["timezone_modifier$ebnf$1"],
    postprocess: ([args]) => args.join(""),
  },
  { name: "prefix", symbols: [/[+\-]/], postprocess: id },
  { name: "num$ebnf$1", symbols: [/[0-9]/] },
  {
    name: "num$ebnf$1",
    symbols: ["num$ebnf$1", /[0-9]/],
    postprocess: function arrpush(d) {
      return d[0].concat([d[1]]);
    },
  },
  {
    name: "num",
    symbols: ["num$ebnf$1"],
    postprocess: ([args]) => Number(args.join("")),
  },
  { name: "grain", symbols: [/[sSmhHdDwWqQMyY]/], postprocess: id },
];
let ParserStart = "rill_time";
export default { Lexer, ParserRules, ParserStart };
