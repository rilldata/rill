<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import CheckboxCard from "@rilldata/web-common/components/forms/CheckboxCard.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import Textarea from "@rilldata/web-common/components/forms/Textarea.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import type { V1Message } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import {
    appendEvalCase,
    createEvalFileWithCase,
    draftEvalCaseFromExchange,
    listEvalFiles,
    type EvalCaseDraft,
  } from "./draft-eval-case";

  export let open: boolean;
  export let messages: V1Message[] = [];
  export let targetMessageId: string | null = null;
  // When set, the dialog uses this draft instead of extracting one from the messages.
  export let prefill: EvalCaseDraft | null = null;
  export let onClose: () => void;
  // Navigating to the file is surface-specific (web-local editor vs cloud), so the host provides it.
  export let onSaved: ((filePath: string) => void) | undefined = undefined;

  const client = useRuntimeClient();
  const NEW_FILE = "__new__";
  const caseNameRegexp = /^[a-zA-Z0-9_-]+$/;

  let draft: EvalCaseDraft | null = null;
  let targetFile = NEW_FILE;
  let includeQueryExpectations = true;
  let saving = false;
  let saveError: string | null = null;

  $: evalFiles = open ? listEvalFiles() : [];
  $: fileOptions = [
    ...evalFiles.map((f) => ({ value: f.path, label: f.name })),
    { value: NEW_FILE, label: m.add_to_eval_new_file() },
  ];

  // Snapshot the draft only when the dialog opens for a target, never while it's open:
  // the messages array gets a new identity on every streamed frame, and re-drafting
  // from it would wipe the user's edits mid-typing.
  // initDraft is a plain function so its reads don't become reactive dependencies.
  let initializedFor: string | null = null;
  $: currentTarget = open
    ? `${prefill ? "prefill" : "message"}:${targetMessageId ?? ""}`
    : null;
  $: if (currentTarget !== initializedFor) {
    initializedFor = currentTarget;
    if (currentTarget !== null) initDraft();
  }

  function initDraft() {
    draft = prefill
      ? { ...prefill }
      : draftEvalCaseFromExchange(messages, targetMessageId ?? "");
    targetFile = listEvalFiles()[0]?.path ?? NEW_FILE;
    includeQueryExpectations = true;
    saveError = null;
  }

  $: hasQueryExpectations = !!(
    draft?.metricsView ||
    draft?.measures?.length ||
    draft?.dimensions?.length
  );

  // Mirror the parser's validation so a saved case can never break the target file.
  $: nameInvalid = !!draft && !caseNameRegexp.test(draft.name ?? "");
  $: missingExpectation =
    !!draft &&
    !draft.expectedAnswer?.trim() &&
    !(includeQueryExpectations && hasQueryExpectations);
  $: saveDisabled =
    !draft || !draft.question?.trim() || nameInvalid || missingExpectation;

  async function handleSave() {
    if (!draft || saving || saveDisabled) return;
    saving = true;
    saveError = null;

    const toSave: EvalCaseDraft = includeQueryExpectations
      ? draft
      : {
          name: draft.name,
          question: draft.question,
          expectedAnswer: draft.expectedAnswer,
          notes: draft.notes,
        };

    try {
      let filePath = targetFile;
      if (targetFile === NEW_FILE) {
        filePath = await createEvalFileWithCase(client, toSave);
      } else {
        await appendEvalCase(client, targetFile, toSave);
      }
      handleClose();
      onSaved?.(filePath);
    } catch (e) {
      saveError = e instanceof Error ? e.message : String(e);
    } finally {
      saving = false;
    }
  }

  function handleClose() {
    onClose();
    draft = null;
  }
</script>

<Dialog.Root
  {open}
  onOpenChange={(isOpen) => {
    if (!isOpen) handleClose();
  }}
>
  <Dialog.Content class="max-w-lg">
    <Dialog.Header>
      <Dialog.Title>{m.add_to_eval_title()}</Dialog.Title>
      <Dialog.Description>{m.add_to_eval_description()}</Dialog.Description>
    </Dialog.Header>

    {#if draft}
      <div class="eval-form">
        <Select
          id="add-to-eval-file"
          label={m.add_to_eval_target_file()}
          options={fileOptions}
          bind:value={targetFile}
        />
        <Input
          id="add-to-eval-name"
          label={m.add_to_eval_case_name()}
          bind:value={draft.name}
        />
        {#if nameInvalid}
          <p class="field-error">{m.add_to_eval_invalid_name()}</p>
        {/if}
        <Textarea
          id="add-to-eval-question"
          label={m.add_to_eval_question()}
          bind:value={draft.question}
          rows={2}
        />
        <Textarea
          id="add-to-eval-answer"
          label={m.add_to_eval_expected_answer()}
          bind:value={draft.expectedAnswer}
          rows={4}
        />
        {#if hasQueryExpectations}
          <CheckboxCard
            checked={includeQueryExpectations}
            label={m.add_to_eval_include_query()}
            onCheckedChange={() =>
              (includeQueryExpectations = !includeQueryExpectations)}
          />
        {/if}
        {#if missingExpectation}
          <p class="field-error">{m.add_to_eval_missing_expectation()}</p>
        {/if}
        {#if saveError}
          <p class="save-error">{saveError}</p>
        {/if}
      </div>
    {/if}

    <Dialog.Footer class="gap-x-2">
      <Button type="secondary" onClick={handleClose}>
        {m.common_cancel()}
      </Button>
      <Button
        type="primary"
        onClick={handleSave}
        disabled={saveDisabled || saving}
        loading={saving}
      >
        {m.add_to_eval_save()}
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>

<style lang="postcss">
  .eval-form {
    @apply flex flex-col gap-3 py-4;
  }

  .save-error {
    @apply text-sm text-red-600;
  }

  .field-error {
    @apply -mt-2 text-xs text-red-600;
  }
</style>
