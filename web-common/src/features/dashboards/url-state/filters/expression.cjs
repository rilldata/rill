// Generated automatically by nearley, version 2.20.1
// http://github.com/Hardmath123/nearley
function id(x) { return x[0]; }

  import {
    binaryPostprocessor,
    inPostprocessor,
    havingPostprocessor,
    andOrPostprocessor,
  } from "./post-processors.ts";
let Lexer = undefined;
let ParserRules = [
    {"name": "_$ebnf$1", "symbols": []},
    {"name": "_$ebnf$1", "symbols": ["_$ebnf$1", "wschar"], "postprocess": function arrpush(d) {return d[0].concat([d[1]]);}},
    {"name": "_", "symbols": ["_$ebnf$1"], "postprocess": function(d) {return null;}},
    {"name": "__$ebnf$1", "symbols": ["wschar"]},
    {"name": "__$ebnf$1", "symbols": ["__$ebnf$1", "wschar"], "postprocess": function arrpush(d) {return d[0].concat([d[1]]);}},
    {"name": "__", "symbols": ["__$ebnf$1"], "postprocess": function(d) {return null;}},
    {"name": "wschar", "symbols": [/[ \t\n\v\f]/], "postprocess": id},
    {"name": "unsigned_int$ebnf$1", "symbols": [/[0-9]/]},
    {"name": "unsigned_int$ebnf$1", "symbols": ["unsigned_int$ebnf$1", /[0-9]/], "postprocess": function arrpush(d) {return d[0].concat([d[1]]);}},
    {"name": "unsigned_int", "symbols": ["unsigned_int$ebnf$1"], "postprocess": 
        function(d) {
            return parseInt(d[0].join(""));
        }
        },
    {"name": "int$ebnf$1$subexpression$1", "symbols": [{"literal":"-"}]},
    {"name": "int$ebnf$1$subexpression$1", "symbols": [{"literal":"+"}]},
    {"name": "int$ebnf$1", "symbols": ["int$ebnf$1$subexpression$1"], "postprocess": id},
    {"name": "int$ebnf$1", "symbols": [], "postprocess": function(d) {return null;}},
    {"name": "int$ebnf$2", "symbols": [/[0-9]/]},
    {"name": "int$ebnf$2", "symbols": ["int$ebnf$2", /[0-9]/], "postprocess": function arrpush(d) {return d[0].concat([d[1]]);}},
    {"name": "int", "symbols": ["int$ebnf$1", "int$ebnf$2"], "postprocess": 
        function(d) {
            if (d[0]) {
                return parseInt(d[0][0]+d[1].join(""));
            } else {
                return parseInt(d[1].join(""));
            }
        }
        },
    {"name": "unsigned_decimal$ebnf$1", "symbols": [/[0-9]/]},
    {"name": "unsigned_decimal$ebnf$1", "symbols": ["unsigned_decimal$ebnf$1", /[0-9]/], "postprocess": function arrpush(d) {return d[0].concat([d[1]]);}},
    {"name": "unsigned_decimal$ebnf$2$subexpression$1$ebnf$1", "symbols": [/[0-9]/]},
    {"name": "unsigned_decimal$ebnf$2$subexpression$1$ebnf$1", "symbols": ["unsigned_decimal$ebnf$2$subexpression$1$ebnf$1", /[0-9]/], "postprocess": function arrpush(d) {return d[0].concat([d[1]]);}},
    {"name": "unsigned_decimal$ebnf$2$subexpression$1", "symbols": [{"literal":"."}, "unsigned_decimal$ebnf$2$subexpression$1$ebnf$1"]},
    {"name": "unsigned_decimal$ebnf$2", "symbols": ["unsigned_decimal$ebnf$2$subexpression$1"], "postprocess": id},
    {"name": "unsigned_decimal$ebnf$2", "symbols": [], "postprocess": function(d) {return null;}},
    {"name": "unsigned_decimal", "symbols": ["unsigned_decimal$ebnf$1", "unsigned_decimal$ebnf$2"], "postprocess": 
        function(d) {
            return parseFloat(
                d[0].join("") +
                (d[1] ? "."+d[1][1].join("") : "")
            );
        }
        },
    {"name": "decimal$ebnf$1", "symbols": [{"literal":"-"}], "postprocess": id},
    {"name": "decimal$ebnf$1", "symbols": [], "postprocess": function(d) {return null;}},
    {"name": "decimal$ebnf$2", "symbols": [/[0-9]/]},
    {"name": "decimal$ebnf$2", "symbols": ["decimal$ebnf$2", /[0-9]/], "postprocess": function arrpush(d) {return d[0].concat([d[1]]);}},
    {"name": "decimal$ebnf$3$subexpression$1$ebnf$1", "symbols": [/[0-9]/]},
    {"name": "decimal$ebnf$3$subexpression$1$ebnf$1", "symbols": ["decimal$ebnf$3$subexpression$1$ebnf$1", /[0-9]/], "postprocess": function arrpush(d) {return d[0].concat([d[1]]);}},
    {"name": "decimal$ebnf$3$subexpression$1", "symbols": [{"literal":"."}, "decimal$ebnf$3$subexpression$1$ebnf$1"]},
    {"name": "decimal$ebnf$3", "symbols": ["decimal$ebnf$3$subexpression$1"], "postprocess": id},
    {"name": "decimal$ebnf$3", "symbols": [], "postprocess": function(d) {return null;}},
    {"name": "decimal", "symbols": ["decimal$ebnf$1", "decimal$ebnf$2", "decimal$ebnf$3"], "postprocess": 
        function(d) {
            return parseFloat(
                (d[0] || "") +
                d[1].join("") +
                (d[2] ? "."+d[2][1].join("") : "")
            );
        }
        },
    {"name": "percentage", "symbols": ["decimal", {"literal":"%"}], "postprocess": 
        function(d) {
            return d[0]/100;
        }
        },
    {"name": "jsonfloat$ebnf$1", "symbols": [{"literal":"-"}], "postprocess": id},
    {"name": "jsonfloat$ebnf$1", "symbols": [], "postprocess": function(d) {return null;}},
    {"name": "jsonfloat$ebnf$2", "symbols": [/[0-9]/]},
    {"name": "jsonfloat$ebnf$2", "symbols": ["jsonfloat$ebnf$2", /[0-9]/], "postprocess": function arrpush(d) {return d[0].concat([d[1]]);}},
    {"name": "jsonfloat$ebnf$3$subexpression$1$ebnf$1", "symbols": [/[0-9]/]},
    {"name": "jsonfloat$ebnf$3$subexpression$1$ebnf$1", "symbols": ["jsonfloat$ebnf$3$subexpression$1$ebnf$1", /[0-9]/], "postprocess": function arrpush(d) {return d[0].concat([d[1]]);}},
    {"name": "jsonfloat$ebnf$3$subexpression$1", "symbols": [{"literal":"."}, "jsonfloat$ebnf$3$subexpression$1$ebnf$1"]},
    {"name": "jsonfloat$ebnf$3", "symbols": ["jsonfloat$ebnf$3$subexpression$1"], "postprocess": id},
    {"name": "jsonfloat$ebnf$3", "symbols": [], "postprocess": function(d) {return null;}},
    {"name": "jsonfloat$ebnf$4$subexpression$1$ebnf$1", "symbols": [/[+-]/], "postprocess": id},
    {"name": "jsonfloat$ebnf$4$subexpression$1$ebnf$1", "symbols": [], "postprocess": function(d) {return null;}},
    {"name": "jsonfloat$ebnf$4$subexpression$1$ebnf$2", "symbols": [/[0-9]/]},
    {"name": "jsonfloat$ebnf$4$subexpression$1$ebnf$2", "symbols": ["jsonfloat$ebnf$4$subexpression$1$ebnf$2", /[0-9]/], "postprocess": function arrpush(d) {return d[0].concat([d[1]]);}},
    {"name": "jsonfloat$ebnf$4$subexpression$1", "symbols": [/[eE]/, "jsonfloat$ebnf$4$subexpression$1$ebnf$1", "jsonfloat$ebnf$4$subexpression$1$ebnf$2"]},
    {"name": "jsonfloat$ebnf$4", "symbols": ["jsonfloat$ebnf$4$subexpression$1"], "postprocess": id},
    {"name": "jsonfloat$ebnf$4", "symbols": [], "postprocess": function(d) {return null;}},
    {"name": "jsonfloat", "symbols": ["jsonfloat$ebnf$1", "jsonfloat$ebnf$2", "jsonfloat$ebnf$3", "jsonfloat$ebnf$4"], "postprocess": 
        function(d) {
            return parseFloat(
                (d[0] || "") +
                d[1].join("") +
                (d[2] ? "."+d[2][1].join("") : "") +
                (d[3] ? "e" + (d[3][1] || "+") + d[3][2].join("") : "")
            );
        }
        },
    {"name": "dqstring$ebnf$1", "symbols": []},
    {"name": "dqstring$ebnf$1", "symbols": ["dqstring$ebnf$1", "dstrchar"], "postprocess": function arrpush(d) {return d[0].concat([d[1]]);}},
    {"name": "dqstring", "symbols": [{"literal":"\""}, "dqstring$ebnf$1", {"literal":"\""}], "postprocess": function(d) {return d[1].join(""); }},
    {"name": "sqstring$ebnf$1", "symbols": []},
    {"name": "sqstring$ebnf$1", "symbols": ["sqstring$ebnf$1", "sstrchar"], "postprocess": function arrpush(d) {return d[0].concat([d[1]]);}},
    {"name": "sqstring", "symbols": [{"literal":"'"}, "sqstring$ebnf$1", {"literal":"'"}], "postprocess": function(d) {return d[1].join(""); }},
    {"name": "btstring$ebnf$1", "symbols": []},
    {"name": "btstring$ebnf$1", "symbols": ["btstring$ebnf$1", /[^`]/], "postprocess": function arrpush(d) {return d[0].concat([d[1]]);}},
    {"name": "btstring", "symbols": [{"literal":"`"}, "btstring$ebnf$1", {"literal":"`"}], "postprocess": function(d) {return d[1].join(""); }},
    {"name": "dstrchar", "symbols": [/[^\\"\n]/], "postprocess": id},
    {"name": "dstrchar", "symbols": [{"literal":"\\"}, "strescape"], "postprocess": 
        function(d) {
            return JSON.parse("\""+d.join("")+"\"");
        }
        },
    {"name": "sstrchar", "symbols": [/[^\\'\n]/], "postprocess": id},
    {"name": "sstrchar", "symbols": [{"literal":"\\"}, "strescape"], "postprocess": function(d) { return JSON.parse("\""+d.join("")+"\""); }},
    {"name": "sstrchar$string$1", "symbols": [{"literal":"\\"}, {"literal":"'"}], "postprocess": function joiner(d) {return d.join('');}},
    {"name": "sstrchar", "symbols": ["sstrchar$string$1"], "postprocess": function(d) {return "'"; }},
    {"name": "strescape", "symbols": [/["\\/bfnrt]/], "postprocess": id},
    {"name": "strescape", "symbols": [{"literal":"u"}, /[a-fA-F0-9]/, /[a-fA-F0-9]/, /[a-fA-F0-9]/, /[a-fA-F0-9]/], "postprocess": 
        function(d) {
            return d.join("");
        }
        },
    {"name": "expr", "symbols": ["boolean_expr"], "postprocess": id},
    {"name": "expr$ebnf$1$subexpression$1$subexpression$1", "symbols": [/[aA]/, /[nN]/, /[dD]/], "postprocess": function(d) {return d.join(""); }},
    {"name": "expr$ebnf$1$subexpression$1", "symbols": ["boolean_expr", "_", "expr$ebnf$1$subexpression$1$subexpression$1", "_"]},
    {"name": "expr$ebnf$1", "symbols": ["expr$ebnf$1$subexpression$1"]},
    {"name": "expr$ebnf$1$subexpression$2$subexpression$1", "symbols": [/[aA]/, /[nN]/, /[dD]/], "postprocess": function(d) {return d.join(""); }},
    {"name": "expr$ebnf$1$subexpression$2", "symbols": ["boolean_expr", "_", "expr$ebnf$1$subexpression$2$subexpression$1", "_"]},
    {"name": "expr$ebnf$1", "symbols": ["expr$ebnf$1", "expr$ebnf$1$subexpression$2"], "postprocess": function arrpush(d) {return d[0].concat([d[1]]);}},
    {"name": "expr", "symbols": ["expr$ebnf$1", "non_and_expr"], "postprocess": andOrPostprocessor},
    {"name": "expr$ebnf$2$subexpression$1$subexpression$1", "symbols": [/[oO]/, /[rR]/], "postprocess": function(d) {return d.join(""); }},
    {"name": "expr$ebnf$2$subexpression$1", "symbols": ["boolean_expr", "_", "expr$ebnf$2$subexpression$1$subexpression$1", "_"]},
    {"name": "expr$ebnf$2", "symbols": ["expr$ebnf$2$subexpression$1"]},
    {"name": "expr$ebnf$2$subexpression$2$subexpression$1", "symbols": [/[oO]/, /[rR]/], "postprocess": function(d) {return d.join(""); }},
    {"name": "expr$ebnf$2$subexpression$2", "symbols": ["boolean_expr", "_", "expr$ebnf$2$subexpression$2$subexpression$1", "_"]},
    {"name": "expr$ebnf$2", "symbols": ["expr$ebnf$2", "expr$ebnf$2$subexpression$2"], "postprocess": function arrpush(d) {return d[0].concat([d[1]]);}},
    {"name": "expr", "symbols": ["expr$ebnf$2", "non_or_expr"], "postprocess": andOrPostprocessor},
    {"name": "non_and_expr", "symbols": ["boolean_expr"], "postprocess": id},
    {"name": "non_and_expr$ebnf$1$subexpression$1$subexpression$1", "symbols": [/[oO]/, /[rR]/], "postprocess": function(d) {return d.join(""); }},
    {"name": "non_and_expr$ebnf$1$subexpression$1", "symbols": ["boolean_expr", "__", "non_and_expr$ebnf$1$subexpression$1$subexpression$1", "__"]},
    {"name": "non_and_expr$ebnf$1", "symbols": ["non_and_expr$ebnf$1$subexpression$1"]},
    {"name": "non_and_expr$ebnf$1$subexpression$2$subexpression$1", "symbols": [/[oO]/, /[rR]/], "postprocess": function(d) {return d.join(""); }},
    {"name": "non_and_expr$ebnf$1$subexpression$2", "symbols": ["boolean_expr", "__", "non_and_expr$ebnf$1$subexpression$2$subexpression$1", "__"]},
    {"name": "non_and_expr$ebnf$1", "symbols": ["non_and_expr$ebnf$1", "non_and_expr$ebnf$1$subexpression$2"], "postprocess": function arrpush(d) {return d[0].concat([d[1]]);}},
    {"name": "non_and_expr", "symbols": ["non_and_expr$ebnf$1", "non_and_expr"], "postprocess": andOrPostprocessor},
    {"name": "non_or_expr", "symbols": ["boolean_expr"], "postprocess": id},
    {"name": "non_or_expr$ebnf$1$subexpression$1$subexpression$1", "symbols": [/[aA]/, /[nN]/, /[dD]/], "postprocess": function(d) {return d.join(""); }},
    {"name": "non_or_expr$ebnf$1$subexpression$1", "symbols": ["boolean_expr", "__", "non_or_expr$ebnf$1$subexpression$1$subexpression$1", "__"]},
    {"name": "non_or_expr$ebnf$1", "symbols": ["non_or_expr$ebnf$1$subexpression$1"]},
    {"name": "non_or_expr$ebnf$1$subexpression$2$subexpression$1", "symbols": [/[aA]/, /[nN]/, /[dD]/], "postprocess": function(d) {return d.join(""); }},
    {"name": "non_or_expr$ebnf$1$subexpression$2", "symbols": ["boolean_expr", "__", "non_or_expr$ebnf$1$subexpression$2$subexpression$1", "__"]},
    {"name": "non_or_expr$ebnf$1", "symbols": ["non_or_expr$ebnf$1", "non_or_expr$ebnf$1$subexpression$2"], "postprocess": function arrpush(d) {return d[0].concat([d[1]]);}},
    {"name": "non_or_expr", "symbols": ["non_or_expr$ebnf$1", "non_or_expr"], "postprocess": andOrPostprocessor},
    {"name": "boolean_expr", "symbols": [{"literal":"("}, "expr", {"literal":")"}], "postprocess": ([_, expr]) => expr},
    {"name": "boolean_expr", "symbols": ["column", "__", "in_operator", "_", {"literal":"("}, "value_list", {"literal":")"}], "postprocess": inPostprocessor},
    {"name": "boolean_expr$subexpression$1", "symbols": [/[hH]/, /[aA]/, /[vV]/, /[iI]/, /[nN]/, /[gG]/], "postprocess": function(d) {return d.join(""); }},
    {"name": "boolean_expr", "symbols": ["column", "__", "boolean_expr$subexpression$1", "_", {"literal":"("}, "expr", {"literal":")"}], "postprocess": havingPostprocessor},
    {"name": "boolean_expr", "symbols": ["simple_expr", "_", "compare_operator", "_", "value"], "postprocess": binaryPostprocessor},
    {"name": "simple_expr", "symbols": ["column"], "postprocess": id},
    {"name": "simple_expr", "symbols": ["value"], "postprocess": id},
    {"name": "in_operator$subexpression$1", "symbols": [/[iI]/, /[nN]/], "postprocess": function(d) {return d.join(""); }},
    {"name": "in_operator", "symbols": ["in_operator$subexpression$1"], "postprocess": id},
    {"name": "in_operator$subexpression$2", "symbols": [/[nN]/, /[iI]/, /[nN]/], "postprocess": function(d) {return d.join(""); }},
    {"name": "in_operator", "symbols": ["in_operator$subexpression$2"], "postprocess": id},
    {"name": "in_operator$subexpression$3", "symbols": [/[nN]/, /[oO]/, /[tT]/, {"literal":" "}, /[iI]/, /[nN]/], "postprocess": function(d) {return d.join(""); }},
    {"name": "in_operator", "symbols": ["in_operator$subexpression$3"], "postprocess": () => "NIN"},
    {"name": "compare_operator$subexpression$1", "symbols": [/[eE]/, /[qQ]/], "postprocess": function(d) {return d.join(""); }},
    {"name": "compare_operator", "symbols": ["compare_operator$subexpression$1"], "postprocess": id},
    {"name": "compare_operator$subexpression$2", "symbols": [/[nN]/, /[eE]/, /[qQ]/], "postprocess": function(d) {return d.join(""); }},
    {"name": "compare_operator", "symbols": ["compare_operator$subexpression$2"], "postprocess": id},
    {"name": "compare_operator$subexpression$3", "symbols": [/[gG]/, /[tT]/], "postprocess": function(d) {return d.join(""); }},
    {"name": "compare_operator", "symbols": ["compare_operator$subexpression$3"], "postprocess": id},
    {"name": "compare_operator$subexpression$4", "symbols": [/[gG]/, /[tT]/, /[eE]/], "postprocess": function(d) {return d.join(""); }},
    {"name": "compare_operator", "symbols": ["compare_operator$subexpression$4"], "postprocess": id},
    {"name": "compare_operator$subexpression$5", "symbols": [/[lL]/, /[tT]/], "postprocess": function(d) {return d.join(""); }},
    {"name": "compare_operator", "symbols": ["compare_operator$subexpression$5"], "postprocess": id},
    {"name": "compare_operator$subexpression$6", "symbols": [/[lL]/, /[tT]/, /[eE]/], "postprocess": function(d) {return d.join(""); }},
    {"name": "compare_operator", "symbols": ["compare_operator$subexpression$6"], "postprocess": id},
    {"name": "column", "symbols": ["sqstring"], "postprocess": id},
    {"name": "column$ebnf$1", "symbols": []},
    {"name": "column$ebnf$1", "symbols": ["column$ebnf$1", /[a-zA-Z0-9_]/], "postprocess": function arrpush(d) {return d[0].concat([d[1]]);}},
    {"name": "column", "symbols": [/[a-zA-Z]/, "column$ebnf$1"], "postprocess": ([fst, rest]) => [fst, ...rest].join("")},
    {"name": "value", "symbols": ["sqstring"], "postprocess": id},
    {"name": "value", "symbols": ["int"], "postprocess": id},
    {"name": "value", "symbols": ["decimal"], "postprocess": id},
    {"name": "value_list", "symbols": ["value_list", "_", {"literal":","}, "_", "value"], "postprocess": ([list, _1, _2, _3, value]) => [...list, value]},
    {"name": "value_list", "symbols": ["value"], "postprocess": ([v]) => [v]}
];
let ParserStart = "expr";
export default { Lexer, ParserRules, ParserStart };
