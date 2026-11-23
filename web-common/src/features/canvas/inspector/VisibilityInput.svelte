<script lang="ts">
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import type { Document } from "yaml";
  import { isHiddenPath, type YAMLPath } from "../layout-util";
  import { setKeyBeforeContent } from "../stores/block-edit";
  import type { CanvasEntity } from "../stores/canvas-entity";

  // The "Visible when" input for the `if` condition of a canvas element: a component's
  // item, a tab group or a container.
  export let canvas: CanvasEntity;
  // The element's YAML path, e.g. ["rows", 0, "items", 1] or ["rows", 2].
  export let path: YAMLPath;
  // Who sees the element, worded for its kind.
  export let hint: string;
  // Shown while a condition is set, e.g. that hiding a component doesn't restrict its data.
  export let dataNote: string | undefined = undefined;
  export let id = "canvas-visibility-condition";

  $: ({ parsedContent, hiddenPaths } = canvas);

  $: condition = readCondition($parsedContent, path);

  let value = "";
  $: value = condition;

  $: hiddenFromYou = isHiddenPath($hiddenPaths, path);

  function readCondition(doc: Document | undefined, path: YAMLPath): string {
    const raw: unknown = doc?.getIn([...path, "if"]);
    if (typeof raw === "string") return raw;
    if (typeof raw === "boolean") return String(raw);
    return "";
  }

  function save() {
    const fileArtifact = canvas.fileArtifact;
    if (!fileArtifact || !$parsedContent) return;

    const next = value.trim();
    if (next === condition) return;

    if (next === "") {
      $parsedContent.deleteIn([...path, "if"]);
    } else {
      setKeyBeforeContent($parsedContent, path, "if", next);
    }
    fileArtifact.updateEditorContent($parsedContent.toString(), false, true);
  }
</script>

<div class="visibility">
  <Input
    {id}
    capitalizeLabel={false}
    textClass="text-sm"
    size="sm"
    labelGap={2}
    fontFamily="var(--font-mono, monospace)"
    label={m.canvas_visibility_label()}
    {hint}
    bind:value
    onBlur={save}
    onEnter={save}
  />
  {#if hiddenFromYou}
    <p class="note">{m.canvas_visibility_hidden_from_you()}</p>
  {/if}
  {#if condition && dataNote}
    <p class="note">{dataNote}</p>
  {/if}
</div>

<style lang="postcss">
  .visibility {
    @apply flex flex-col gap-y-1 px-5 py-3 border-t;
  }

  .note {
    @apply text-xs text-fg-secondary;
  }
</style>
