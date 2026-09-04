import { describe, expect, it } from "vitest";
import { V1AIEvalVerdict } from "@rilldata/web-common/runtime-client";
import { computeRunDelta } from "./run-delta";

const PASS = V1AIEvalVerdict.AI_EVAL_VERDICT_PASS;
const FAIL = V1AIEvalVerdict.AI_EVAL_VERDICT_FAIL;
const ERROR = V1AIEvalVerdict.AI_EVAL_VERDICT_ERROR;

function execution(verdicts: Record<string, V1AIEvalVerdict>) {
  return {
    caseResults: Object.entries(verdicts).map(([caseName, verdict]) => ({
      caseName,
      verdict,
    })),
  };
}

describe("computeRunDelta", () => {
  it("classifies fixed, regressed and still-failing cases", () => {
    const previous = execution({ a: FAIL, b: PASS, c: FAIL, d: PASS });
    const latest = execution({ a: PASS, b: FAIL, c: FAIL, d: PASS });
    const delta = computeRunDelta(latest, previous);
    expect(delta.fixed).toEqual(["a"]);
    expect(delta.regressed).toEqual(["b"]);
    expect(delta.stillFailing).toEqual(["c"]);
  });

  it("ignores infra errors and cases missing from either run", () => {
    const previous = execution({ a: FAIL, b: ERROR });
    const latest = execution({ a: ERROR, b: PASS, newcase: FAIL });
    const delta = computeRunDelta(latest, previous);
    expect(delta.fixed).toEqual([]);
    expect(delta.regressed).toEqual([]);
    expect(delta.stillFailing).toEqual([]);
  });

  it("returns an empty delta without two runs", () => {
    const delta = computeRunDelta(execution({ a: PASS }), undefined);
    expect(delta).toEqual({ fixed: [], regressed: [], stillFailing: [] });
  });
});
