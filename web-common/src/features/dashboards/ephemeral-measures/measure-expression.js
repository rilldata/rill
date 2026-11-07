// Generated automatically by nearley, version 2.20.1
// http://github.com/Hardmath123/nearley
function id(x) {
  return x[0];
}
let Lexer = undefined;
let ParserRules = [
  {
    name: "main",
    symbols: ["_", "additive", "_"],
    postprocess: ([, expr]) => expr,
  },
  {
    name: "additive",
    symbols: ["additive", "_", /[+\-]/, "_", "multiplicative"],
    postprocess: ([left, , op, , right], pos) => ({
      type: "binary",
      op,
      left,
      right,
      pos,
    }),
  },
  { name: "additive", symbols: ["multiplicative"], postprocess: id },
  {
    name: "multiplicative",
    symbols: ["multiplicative", "_", /[*/%]/, "_", "unary"],
    postprocess: ([left, , op, , right], pos) => ({
      type: "binary",
      op,
      left,
      right,
      pos,
    }),
  },
  { name: "multiplicative", symbols: ["unary"], postprocess: id },
  {
    name: "unary",
    symbols: [{ literal: "-" }, "_", "unary"],
    postprocess: ([, , expr], pos) => ({ type: "unary", expr, pos }),
  },
  { name: "unary", symbols: ["primary"], postprocess: id },
  {
    name: "primary",
    symbols: ["number"],
    postprocess: ([literal], pos) => ({ type: "literal", literal, pos }),
  },
  {
    name: "primary",
    symbols: ["quoted_ident"],
    postprocess: ([name], pos) => ({ type: "ref", name, pos }),
  },
  { name: "primary$ebnf$1", symbols: ["args"], postprocess: id },
  {
    name: "primary$ebnf$1",
    symbols: [],
    postprocess: function (d) {
      return null;
    },
  },
  {
    name: "primary",
    symbols: [
      "ident",
      "_",
      { literal: "(" },
      "_",
      "primary$ebnf$1",
      "_",
      { literal: ")" },
    ],
    postprocess: ([name, , , , args], pos) => ({
      type: "func",
      name,
      args: args ?? [],
      pos,
    }),
  },
  {
    name: "primary",
    symbols: ["ident"],
    postprocess: ([name], pos) => ({ type: "word", name, pos }),
  },
  {
    name: "primary",
    symbols: [{ literal: "(" }, "_", "additive", "_", { literal: ")" }],
    postprocess: ([, , expr], pos) => ({ type: "paren", expr, pos }),
  },
  { name: "args$ebnf$1", symbols: [] },
  {
    name: "args$ebnf$1$subexpression$1",
    symbols: ["_", { literal: "," }, "_", "additive"],
  },
  {
    name: "args$ebnf$1",
    symbols: ["args$ebnf$1", "args$ebnf$1$subexpression$1"],
    postprocess: function arrpush(d) {
      return d[0].concat([d[1]]);
    },
  },
  {
    name: "args",
    symbols: ["additive", "args$ebnf$1"],
    postprocess: ([first, rest]) => [first, ...rest.map(([, , , arg]) => arg)],
  },
  { name: "ident$ebnf$1", symbols: [] },
  {
    name: "ident$ebnf$1",
    symbols: ["ident$ebnf$1", /[A-Za-z0-9_]/],
    postprocess: function arrpush(d) {
      return d[0].concat([d[1]]);
    },
  },
  {
    name: "ident",
    symbols: [/[A-Za-z_]/, "ident$ebnf$1"],
    postprocess: ([first, rest]) => first + rest.join(""),
  },
  { name: "quoted_ident$ebnf$1", symbols: ["qchar"] },
  {
    name: "quoted_ident$ebnf$1",
    symbols: ["quoted_ident$ebnf$1", "qchar"],
    postprocess: function arrpush(d) {
      return d[0].concat([d[1]]);
    },
  },
  {
    name: "quoted_ident",
    symbols: [{ literal: '"' }, "quoted_ident$ebnf$1", { literal: '"' }],
    postprocess: ([, chars]) => chars.join(""),
  },
  { name: "qchar", symbols: [/[^"]/], postprocess: id },
  {
    name: "qchar$string$1",
    symbols: [{ literal: '"' }, { literal: '"' }],
    postprocess: function joiner(d) {
      return d.join("");
    },
  },
  { name: "qchar", symbols: ["qchar$string$1"], postprocess: () => '"' },
  {
    name: "number$ebnf$1$subexpression$1$ebnf$1",
    symbols: ["digits"],
    postprocess: id,
  },
  {
    name: "number$ebnf$1$subexpression$1$ebnf$1",
    symbols: [],
    postprocess: function (d) {
      return null;
    },
  },
  {
    name: "number$ebnf$1$subexpression$1",
    symbols: [{ literal: "." }, "number$ebnf$1$subexpression$1$ebnf$1"],
  },
  {
    name: "number$ebnf$1",
    symbols: ["number$ebnf$1$subexpression$1"],
    postprocess: id,
  },
  {
    name: "number$ebnf$1",
    symbols: [],
    postprocess: function (d) {
      return null;
    },
  },
  { name: "number$ebnf$2", symbols: ["exponent"], postprocess: id },
  {
    name: "number$ebnf$2",
    symbols: [],
    postprocess: function (d) {
      return null;
    },
  },
  {
    name: "number",
    symbols: ["digits", "number$ebnf$1", "number$ebnf$2"],
    postprocess: ([int, frac, exp]) =>
      int + (frac ? "." + (frac[1] ?? "") : "") + (exp ?? ""),
  },
  { name: "number$ebnf$3", symbols: ["exponent"], postprocess: id },
  {
    name: "number$ebnf$3",
    symbols: [],
    postprocess: function (d) {
      return null;
    },
  },
  {
    name: "number",
    symbols: [{ literal: "." }, "digits", "number$ebnf$3"],
    postprocess: ([, frac, exp]) => "." + frac + (exp ?? ""),
  },
  { name: "digits$ebnf$1", symbols: [/[0-9]/] },
  {
    name: "digits$ebnf$1",
    symbols: ["digits$ebnf$1", /[0-9]/],
    postprocess: function arrpush(d) {
      return d[0].concat([d[1]]);
    },
  },
  {
    name: "digits",
    symbols: ["digits$ebnf$1"],
    postprocess: ([d]) => d.join(""),
  },
  { name: "exponent$ebnf$1", symbols: [/[+\-]/], postprocess: id },
  {
    name: "exponent$ebnf$1",
    symbols: [],
    postprocess: function (d) {
      return null;
    },
  },
  {
    name: "exponent",
    symbols: [/[eE]/, "exponent$ebnf$1", "digits"],
    postprocess: ([e, sign, d]) => e + (sign ?? "") + d,
  },
  { name: "_$ebnf$1", symbols: [] },
  {
    name: "_$ebnf$1",
    symbols: ["_$ebnf$1", /[ \t\n\r]/],
    postprocess: function arrpush(d) {
      return d[0].concat([d[1]]);
    },
  },
  { name: "_", symbols: ["_$ebnf$1"], postprocess: () => null },
];
let ParserStart = "main";
export default { Lexer, ParserRules, ParserStart };
