<!--
  Generates a project-change fix for a feedback item: an ai_instructions rule or
  a new measure definition. The suggestion is shown in an editable textbox with a
  live diff preview of the resulting file change, and the admin chooses to apply
  it, apply and continue editing the file, or discard it and edit manually.
  A draft eval case capturing the conversation can be saved alongside.
-->
<script lang="ts">
  import type { PartialMessage } from "@bufbuild/protobuf";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Diff2HtmlView from "@rilldata/web-common/components/diff/Diff2HtmlView.svelte";
  import Textarea from "@rilldata/web-common/components/forms/Textarea.svelte";
  import AddToEvalDialog from "@rilldata/web-common/features/ai-evals/draft/AddToEvalDialog.svelte";
  import {
    draftEvalCaseFromExchange,
    slugifyEvalCaseName,
    type EvalCaseDraft,
  } from "@rilldata/web-common/features/ai-evals/draft/draft-eval-case";
  import {
    appendAIInstructions,
    appendYAMLMeasure,
  } from "@rilldata/web-common/features/ai-instructions/append-ai-instructions";
  import { navigateToFile } from "@rilldata/web-common/layout/navigation/editor-routing";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import type { GenerateAIFeedbackFixResponse } from "@rilldata/web-common/proto/gen/rill/runtime/v1/api_pb";
  import {
    runtimeServiceGenerateAIFeedbackFix,
    runtimeServiceGetFile,
    runtimeServicePutFile,
    type V1Message,
  } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { createTwoFilesPatch } from "diff";
  import { extractMessageText } from "../core/utils";
  import type { InboxItem } from "./inbox-data";

  export let item: InboxItem;
  export let messages: V1Message[];
  export let onResolve: (status: "open" | "addressed" | "dismissed") => void;

  const client = useRuntimeClient();

  let loading = false;
  let error: string | null = null;
  let result: PartialMessage<GenerateAIFeedbackFixResponse> | null = null;
  // The proposed change payload (ai_instructions rule or measure YAML),
  // editable by the admin before applying.
  let suggestion = "";
  // Current contents of the target file, fetched for the diff preview.
  let currentFileContents: string | null = null;
  let applying = false;
  let applied = false;
  let evalDialogOpen = false;
  let evalPrefill: EvalCaseDraft | null = null;

  $: isMeasure = result?.action === "add_measure";
  $: preview = buildPreview(result, suggestion, currentFileContents);

  async function suggest() {
    loading = true;
    error = null;
    result = null;
    suggestion = "";
    currentFileContents = null;
    applied = false;
    try {
      result = await runtimeServiceGenerateAIFeedbackFix(client, {
        feedback: {
          kind: item.kind,
          sentiment: item.sentiment,
          categories: item.categories,
          comment: item.comment,
          predictedAttribution: item.predictedAttribution,
        },
        transcript: conversationTranscript(messages),
      });
      suggestion =
        (result.action === "add_measure"
          ? result.measureYaml
          : result.instruction) ?? "";
      if (result.filePath) {
        const current = await runtimeServiceGetFile(client, {
          path: result.filePath,
        });
        currentFileContents = current.blob ?? "";
      }
    } catch (e) {
      error = e instanceof Error ? e.message : String(e);
    } finally {
      loading = false;
    }
  }

  // Splices the (possibly tuned) suggestion into the target file. Plain function
  // so callers decide the input; also powers the live diff preview.
  function splice(action: string | undefined, contents: string, text: string) {
    return action === "add_measure"
      ? appendYAMLMeasure(contents, text)
      : appendAIInstructions(contents, text);
  }

  function buildPreview(
    res: typeof result,
    text: string,
    contents: string | null,
  ): { diff: string } | { error: string } | null {
    if (!res?.filePath || contents === null || !text.trim()) return null;
    try {
      const updated = splice(res.action, contents, text.trim());
      return {
        diff: createTwoFilesPatch(res.filePath, res.filePath, contents, updated),
      };
    } catch (e) {
      return { error: e instanceof Error ? e.message : String(e) };
    }
  }

  // Applies the suggestion at apply time against freshly read file contents,
  // so the admin's edits are what lands and concurrent file changes aren't lost.
  async function apply(): Promise<boolean> {
    if (!result?.filePath || !suggestion.trim() || applying) return false;
    applying = true;
    error = null;
    try {
      const current = await runtimeServiceGetFile(client, {
        path: result.filePath,
      });
      const updated = splice(
        result.action,
        current.blob ?? "",
        suggestion.trim(),
      );
      await runtimeServicePutFile(client, {
        path: result.filePath,
        blob: updated,
      });
      applied = true;
      return true;
    } catch (e) {
      error = e instanceof Error ? e.message : String(e);
      return false;
    } finally {
      applying = false;
    }
  }

  async function applyAndEdit() {
    if (!result?.filePath) return;
    const ok = await apply();
    if (ok) await navigateToFile(result.filePath);
  }

  // "Discard and make the change manually": opens the target file without applying.
  async function editManually() {
    if (!result?.filePath) return;
    await navigateToFile(result.filePath);
  }

  function openEvalDialog() {
    if (!result) return;
    const base = draftEvalCaseFromExchange(messages, item.targetMessageId);
    const question = result.evalQuestion || base?.question || "";
    evalPrefill = {
      ...(base ?? {}),
      name: base?.name || slugifyEvalCaseName(question),
      question,
      expectedAnswer: result.evalExpectedAnswer || base?.expectedAnswer || "",
    };
    evalDialogOpen = true;
  }

  // Only the actual conversation (user prompts and assistant answers) goes into the fix prompt;
  // raw tool-result JSON dumps would blow the model context on data-heavy conversations.
  function conversationTranscript(
    msgs: V1Message[],
  ): { role: string; text: string }[] {
    const transcript: { role: string; text: string }[] = [];
    for (const msg of msgs) {
      if (msg.tool !== "router_agent") continue;
      try {
        const content = JSON.parse(msg.contentData ?? "{}");
        if (msg.type === "call" && content?.prompt) {
          transcript.push({ role: "user", text: content.prompt });
        } else if (msg.type === "result" && content?.response) {
          transcript.push({ role: "assistant", text: content.response });
        }
      } catch {
        const text = extractMessageText(msg);
        if (text) transcript.push({ role: msg.role ?? "", text });
      }
    }
    return transcript;
  }
</script>

<div class="fix-card">
  <Button type="secondary" small onClick={suggest} disabled={loading} {loading}>
    {m.suggest_fix()}
  </Button>

  {#if loading}
    <p class="hint">{m.suggest_fix_generating()}</p>
  {/if}
  {#if error}
    <p class="error">{error}</p>
  {/if}

  {#if result}
    <p class="summary">{result.summary}</p>

    {#if result.filePath}
      <p class="target">
        {m.suggest_fix_target_file({ file: result.filePath })}
      </p>
      <Textarea
        id="suggest-fix-suggestion"
        label={isMeasure
          ? m.suggest_fix_measure_label()
          : m.suggest_fix_context_label()}
        bind:value={suggestion}
        rows={isMeasure ? 6 : 4}
        disabled={applied}
      />
      {#if preview && "error" in preview}
        <p class="error">{preview.error}</p>
      {:else if preview && !applied}
        <Diff2HtmlView diff={preview.diff} scrollX />
      {/if}
    {/if}

    <div class="actions">
      {#if result.filePath && !applied}
        <Button
          type="primary"
          small
          onClick={apply}
          disabled={!suggestion.trim() || applying}
          loading={applying}
        >
          {m.suggest_fix_apply()}
        </Button>
        <Button
          type="secondary"
          small
          onClick={applyAndEdit}
          disabled={!suggestion.trim() || applying}
        >
          {m.suggest_fix_apply_and_edit()}
        </Button>
        <Button type="text" small onClick={editManually}>
          {m.suggest_fix_edit_manually()}
        </Button>
      {/if}
      <Button type="secondary" small onClick={openEvalDialog}>
        {m.add_to_eval()}
      </Button>
    </div>

    {#if applied}
      <div class="mark-addressed">
        <span>{m.suggest_fix_mark_addressed_prompt()}</span>
        <Button type="primary" small onClick={() => onResolve("addressed")}>
          {m.feedback_inbox_addressed()}
        </Button>
      </div>
    {/if}
  {/if}
</div>

<AddToEvalDialog
  open={evalDialogOpen}
  prefill={evalPrefill}
  onClose={() => (evalDialogOpen = false)}
  onSaved={() => onResolve("addressed")}
/>

<style lang="postcss">
  .fix-card {
    @apply flex flex-col gap-2 border-t pt-2;
  }

  .hint {
    @apply text-sm text-fg-muted;
  }

  .error {
    @apply text-sm text-red-600;
  }

  .summary {
    @apply text-sm;
  }

  .target {
    @apply text-xs text-fg-muted;
  }

  .actions {
    @apply flex flex-wrap items-center gap-2;
  }

  .mark-addressed {
    @apply flex items-center justify-between gap-2 rounded-md bg-surface-hover px-3 py-2 text-sm;
  }
</style>
