<!--
  Generates AI-context fixes for the latest run's failing eval cases. Each
  proposal's ai_instructions rule is shown in an editable textbox so the admin
  can tune it before applying it to the target file.
-->
<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Textarea from "@rilldata/web-common/components/forms/Textarea.svelte";
  import { appendAIInstructions } from "@rilldata/web-common/features/ai-instructions/append-ai-instructions";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import {
    runtimeServiceGenerateAIEvalFix,
    runtimeServiceGetFile,
    runtimeServicePutFile,
  } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  export let evalName: string;
  // Called after a proposal is applied so the host can offer a re-run.
  export let onApplied: () => void;

  const client = useRuntimeClient();

  // Local editable copy of a generated proposal.
  interface EditableProposal {
    filePath: string;
    reason: string;
    instruction: string;
    applied: boolean;
  }

  let loading = false;
  let error: string | null = null;
  let analysis: string | null = null;
  let proposals: EditableProposal[] = [];

  async function suggest() {
    loading = true;
    error = null;
    analysis = null;
    proposals = [];
    try {
      const result = await runtimeServiceGenerateAIEvalFix(client, {
        eval: evalName,
      });
      analysis = result.analysis ?? "";
      proposals = (result.proposals ?? []).map((p) => ({
        filePath: p.filePath ?? "",
        reason: p.reason ?? "",
        instruction: p.instruction ?? "",
        applied: false,
      }));
    } catch (e) {
      error = e instanceof Error ? e.message : String(e);
    } finally {
      loading = false;
    }
  }

  // Applies the (possibly tuned) instruction by appending it to the target
  // file's ai_instructions at apply time, so the admin's edits are what lands.
  async function apply(proposal: EditableProposal) {
    if (!proposal.filePath || !proposal.instruction.trim()) return;
    error = null;
    try {
      const current = await runtimeServiceGetFile(client, {
        path: proposal.filePath,
      });
      const updated = appendAIInstructions(
        current.blob ?? "",
        proposal.instruction.trim(),
      );
      await runtimeServicePutFile(client, {
        path: proposal.filePath,
        blob: updated,
      });
      proposal.applied = true;
      proposals = proposals;
      onApplied();
    } catch (e) {
      error = e instanceof Error ? e.message : String(e);
    }
  }
</script>

<div class="fix-section">
  <Button type="secondary" small onClick={suggest} disabled={loading} {loading}>
    {m.suggest_fix()}
  </Button>

  {#if loading}
    <p class="hint">{m.suggest_fix_generating()}</p>
  {/if}
  {#if error}
    <p class="error">{error}</p>
  {/if}

  {#if analysis !== null}
    <p class="analysis">{analysis}</p>
    {#each proposals as proposal, i (proposal.filePath)}
      <div class="proposal">
        <span class="proposal-reason">{proposal.reason}</span>
        <p class="target">
          {m.suggest_fix_target_file({ file: proposal.filePath })}
        </p>
        <Textarea
          id="eval-fix-instruction-{i}"
          label={m.suggest_fix_context_label()}
          bind:value={proposal.instruction}
          rows={3}
          disabled={proposal.applied}
        />
        <div class="proposal-actions">
          {#if proposal.applied}
            <span class="applied">{m.suggest_fix_applied()}</span>
          {:else}
            <Button
              type="primary"
              small
              onClick={() => apply(proposal)}
              disabled={!proposal.instruction.trim()}
            >
              {m.suggest_fix_apply()}
            </Button>
          {/if}
        </div>
      </div>
    {/each}
  {/if}
</div>

<style lang="postcss">
  .fix-section {
    @apply flex flex-col gap-2;
  }

  .hint {
    @apply text-sm text-fg-muted;
  }

  .error {
    @apply text-sm text-red-600;
  }

  .analysis {
    @apply text-sm;
  }

  .proposal {
    @apply flex flex-col gap-1 rounded-md border p-2;
  }

  .proposal-reason {
    @apply text-sm text-fg-secondary;
  }

  .target {
    @apply text-xs text-fg-muted;
  }

  .proposal-actions {
    @apply flex items-center gap-2;
  }

  .applied {
    @apply text-sm font-medium text-green-700;
  }
</style>
