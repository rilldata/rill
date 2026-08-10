import {
  V1AIEvalVerdict,
  type V1AIEvalExecution,
} from "@rilldata/web-common/runtime-client";

// RunDelta describes how case verdicts changed between two runs.
export interface RunDelta {
  fixed: string[];
  regressed: string[];
  stillFailing: string[];
}

// computeRunDelta joins two runs' case results by case name and classifies the changes.
// Only PASS and FAIL verdicts participate; ERROR and SKIPPED are infra outcomes, not grades.
export function computeRunDelta(
  latest: V1AIEvalExecution | undefined,
  previous: V1AIEvalExecution | undefined,
): RunDelta {
  const delta: RunDelta = { fixed: [], regressed: [], stillFailing: [] };
  if (!latest || !previous) return delta;

  const previousByCase = new Map(
    (previous.caseResults ?? []).map((cr) => [cr.caseName, cr.verdict]),
  );

  for (const cr of latest.caseResults ?? []) {
    if (!cr.caseName) continue;
    const before = previousByCase.get(cr.caseName);
    const after = cr.verdict;
    if (
      before === V1AIEvalVerdict.AI_EVAL_VERDICT_FAIL &&
      after === V1AIEvalVerdict.AI_EVAL_VERDICT_PASS
    ) {
      delta.fixed.push(cr.caseName);
    } else if (
      before === V1AIEvalVerdict.AI_EVAL_VERDICT_PASS &&
      after === V1AIEvalVerdict.AI_EVAL_VERDICT_FAIL
    ) {
      delta.regressed.push(cr.caseName);
    } else if (
      before === V1AIEvalVerdict.AI_EVAL_VERDICT_FAIL &&
      after === V1AIEvalVerdict.AI_EVAL_VERDICT_FAIL
    ) {
      delta.stillFailing.push(cr.caseName);
    }
  }
  return delta;
}
