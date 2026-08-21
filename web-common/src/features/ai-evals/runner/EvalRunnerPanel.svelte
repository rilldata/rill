<script lang="ts">
  import { page } from "$app/stores";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Checkbox from "@rilldata/web-common/components/forms/Checkbox.svelte";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import type {
    V1AIEvalCaseResult,
    V1AIEvalExecution,
  } from "@rilldata/web-common/runtime-client";
  import { createRuntimeServiceCreateTriggerMutation } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import EvalFixSection from "../fix/EvalFixSection.svelte";
  import { computeRunDelta } from "../fix/run-delta";
  import EvalVerdictChip from "./EvalVerdictChip.svelte";

  export let fileArtifact: FileArtifact;

  const client = useRuntimeClient();
  const triggerMutation = createRuntimeServiceCreateTriggerMutation(client);

  let selectedCases: string[] = [];
  let expandedCase: string | null = null;
  let historyIndex = 0;

  $: ({ resourceName } = fileArtifact);
  $: evalName = $resourceName?.name ?? "";

  $: resourceQuery = fileArtifact.getResource(queryClient);
  $: aiEval = $resourceQuery.data?.aiEval;
  $: cases = aiEval?.spec?.cases ?? [];
  $: currentExecution = aiEval?.state?.currentExecution;
  $: history = aiEval?.state?.executionHistory ?? [];
  $: isRunning = !!currentExecution;

  // Show the in-flight run if there is one, otherwise the selected history entry
  $: displayedExecution = (currentExecution ??
    history[Math.min(historyIndex, Math.max(history.length - 1, 0))]) as
    | V1AIEvalExecution
    | undefined;
  $: resultsByCase = new Map(
    (displayedExecution?.caseResults ?? []).map((cr) => [cr.caseName, cr]),
  );
  // Delta between the latest and previous run, for the fix loop's regression check
  $: latestDelta = computeRunDelta(history[0], history[1]);
  $: hasDelta =
    latestDelta.fixed.length > 0 ||
    latestDelta.regressed.length > 0 ||
    latestDelta.stillFailing.length > 0;
  $: hasFailures = (history[0]?.failedCount ?? 0) > 0;
  let fixApplied = false;

  function toggleCase(name: string) {
    if (selectedCases.includes(name)) {
      selectedCases = selectedCases.filter((c) => c !== name);
    } else {
      selectedCases = [...selectedCases, name];
    }
  }

  function run(caseNames: string[]) {
    historyIndex = 0;
    fixApplied = false;
    $triggerMutation.mutate({
      aiEvals: [{ aiEval: evalName, cases: caseNames }],
    });
  }

  function cancel() {
    $triggerMutation.mutate({
      aiEvals: [{ aiEval: evalName, cancel: true }],
    });
  }

  // Eval transcripts are regular AI conversations viewable at web-local's /ai route.
  // In cloud editing there is no page that loads conversations from the edit deployment's
  // runtime (the -/ai route reads the production runtime), so the link is hidden there.
  $: showTranscriptLinks = !$page.params.organization;

  function transcriptHref(sessionId: string): string {
    return `/ai/${sessionId}`;
  }

  function summaryOf(execution: V1AIEvalExecution): string {
    return m.eval_run_summary({
      passed: String(execution.passedCount ?? 0),
      failed: String(execution.failedCount ?? 0),
      errored: String(execution.erroredCount ?? 0),
    });
  }

  function caseResult(name: string): V1AIEvalCaseResult | undefined {
    return resultsByCase.get(name);
  }
</script>

<div class="runner">
  <div class="section">
    <div class="section-header">
      <h3>{m.eval_cases_title()}</h3>
      <div class="run-actions">
        {#if isRunning}
          <Button type="secondary" small onClick={cancel}>
            {m.eval_cancel()}
          </Button>
        {:else}
          {#if selectedCases.length > 0}
            <Button
              type="secondary"
              small
              onClick={() => run(selectedCases)}
              disabled={!evalName}
            >
              {m.eval_run_selected({ count: String(selectedCases.length) })}
            </Button>
          {/if}
          <Button
            type="primary"
            small
            onClick={() => run([])}
            disabled={!evalName || cases.length === 0}
          >
            {m.eval_run_all()}
          </Button>
        {/if}
      </div>
    </div>

    <div class="case-list">
      {#each cases as c (c.name)}
        {@const result = caseResult(c.name ?? "")}
        <div class="case-row">
          <div class="case-row-main">
            <Checkbox
              inverse
              id={`eval-case-${c.name}`}
              checked={selectedCases.includes(c.name ?? "")}
              onCheckedChange={() => toggleCase(c.name ?? "")}
              disabled={isRunning}
            />
            <button
              class="case-name"
              onclick={() =>
                (expandedCase = expandedCase === c.name ? null : (c.name ?? null))}
            >
              {c.name}
            </button>
            {#if result}
              <EvalVerdictChip verdict={result.verdict} />
            {/if}
          </div>
          {#if expandedCase === c.name}
            <div class="case-detail">
              <p class="detail-question">{c.question}</p>
              {#if c.expectAnswer}
                <div class="detail-block">
                  <span class="detail-label"
                    >{m.add_to_eval_expected_answer()}</span
                  >
                  <p>{c.expectAnswer}</p>
                </div>
              {/if}
              {#if result?.answerPreview}
                <div class="detail-block">
                  <span class="detail-label">{m.eval_actual_answer()}</span>
                  <p>{result.answerPreview}</p>
                </div>
              {/if}
              {#each result?.failureReasons ?? [] as reason (reason)}
                <p class="detail-reason">{reason}</p>
              {/each}
              {#if result?.judgeReasoning}
                <div class="detail-block">
                  <span class="detail-label">{m.eval_judge_reasoning()}</span>
                  <p>{result.judgeReasoning}</p>
                </div>
              {/if}
              {#if result?.sessionId && showTranscriptLinks}
                <a class="transcript-link" href={transcriptHref(result.sessionId)}>
                  {m.eval_open_transcript()}
                </a>
              {/if}
            </div>
          {/if}
        </div>
      {/each}
    </div>
  </div>

  <div class="section">
    <h3>{m.eval_results_title()}</h3>
    {#if isRunning && currentExecution}
      <p class="summary running">{m.eval_running()}</p>
    {:else if !displayedExecution}
      <p class="empty">{m.eval_no_runs()}</p>
    {:else}
      <p class="summary">
        {summaryOf(displayedExecution)}
        {#if displayedExecution.canceled}
          · {m.eval_canceled()}
        {/if}
      </p>
      {#if displayedExecution.errorMessage}
        <p class="detail-reason">{displayedExecution.errorMessage}</p>
      {/if}
    {/if}

    {#if !isRunning && hasDelta}
      <p class="delta">
        {m.eval_fix_delta({
          fixed: String(latestDelta.fixed.length),
          regressed: String(latestDelta.regressed.length),
          stillFailing: String(latestDelta.stillFailing.length),
        })}
      </p>
    {/if}

    {#if history.length > 1}
      <div class="history">
        <span class="detail-label">{m.eval_history()}</span>
        <select bind:value={historyIndex} disabled={isRunning}>
          {#each history as execution, i (i)}
            <option value={i}>
              {execution.startedOn
                ? new Date(execution.startedOn).toLocaleString()
                : "—"}
              — {summaryOf(execution)}
            </option>
          {/each}
        </select>
      </div>
    {/if}
  </div>

  {#if !isRunning && hasFailures}
    <div class="section">
      <EvalFixSection {evalName} onApplied={() => (fixApplied = true)} />
      {#if fixApplied}
        <div class="rerun">
          <Button type="primary" small onClick={() => run([])}>
            {m.eval_fix_rerun()}
          </Button>
        </div>
      {/if}
    </div>
  {/if}
</div>

<style lang="postcss">
  .runner {
    @apply flex h-full w-full flex-col gap-4 overflow-y-auto p-3;
  }

  .section h3 {
    @apply mb-2 text-xs font-semibold uppercase text-fg-muted;
  }

  .section-header {
    @apply flex items-center justify-between gap-2;
  }

  .run-actions {
    @apply flex items-center gap-1;
  }

  .case-list {
    @apply mt-2 flex flex-col gap-1;
  }

  .case-row {
    @apply rounded-md border px-2 py-1.5;
  }

  .case-row-main {
    @apply flex items-center gap-2;
  }

  .case-name {
    @apply flex-1 truncate text-left text-sm font-medium;
  }

  .case-detail {
    @apply mt-2 flex flex-col gap-2 border-t pt-2;
  }

  .detail-question {
    @apply text-sm italic text-fg-secondary;
  }

  .detail-block p {
    @apply text-sm;
  }

  .detail-label {
    @apply text-xs font-semibold uppercase text-fg-muted;
  }

  .detail-reason {
    @apply text-sm text-red-600;
  }

  .transcript-link {
    @apply text-sm text-primary-600 hover:underline;
  }

  .summary {
    @apply text-sm;
  }

  .summary.running {
    @apply text-fg-muted;
  }

  .empty {
    @apply text-sm text-fg-muted;
  }

  .history {
    @apply mt-2 flex flex-col gap-1;
  }

  .history select {
    @apply rounded-md border px-2 py-1 text-sm;
  }

  .delta {
    @apply mt-1 text-sm font-medium;
  }

  .rerun {
    @apply mt-2;
  }
</style>
