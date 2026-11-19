<script lang="ts">
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import type { Document } from "yaml";
  import type { BaseCanvasComponent } from "../components/BaseCanvasComponent";
  import { isHiddenPath } from "../layout-util";

  export let component: BaseCanvasComponent;

  $: ({ parent, specStore } = component);
  $: ({ parsedContent, hiddenPaths, _metricsViews } = parent);

  // The canvas item's YAML path: the component's path without its renderer key.
  $: itemPath = component.pathInYAML.slice(0, -1);

  $: condition = readCondition($parsedContent, itemPath);

  let value = "";
  $: value = condition;

  $: hiddenFromYou = isHiddenPath($hiddenPaths, itemPath);

  $: metricsViewName = ($specStore as { metrics_view?: string } | undefined)
    ?.metrics_view;
  $: dataQueryable =
    !!condition &&
    !!metricsViewName &&
    !$_metricsViews[metricsViewName]?.state?.validSpec?.securityRules?.length;

  function readCondition(
    doc: Document | undefined,
    path: (string | number)[],
  ): string {
    const raw: unknown = doc?.getIn([...path, "if"]);
    if (typeof raw === "string") return raw;
    if (typeof raw === "boolean") return String(raw);
    return "";
  }

  function save() {
    const fileArtifact = parent.fileArtifact;
    if (!fileArtifact || !$parsedContent) return;

    const next = value.trim();
    if (next === condition) return;

    if (next === "") {
      $parsedContent.deleteIn([...itemPath, "if"]);
    } else {
      $parsedContent.setIn([...itemPath, "if"], next);
    }
    fileArtifact.updateEditorContent($parsedContent.toString(), false, true);
  }
</script>

<div class="visibility">
  <Input
    id="canvas-visibility-condition"
    capitalizeLabel={false}
    textClass="text-sm"
    size="sm"
    labelGap={2}
    fontFamily="var(--font-mono, monospace)"
    label={m.canvas_visibility_label()}
    hint={m.canvas_visibility_hint()}
    bind:value
    onBlur={save}
    onEnter={save}
  />
  {#if hiddenFromYou}
    <p class="note">{m.canvas_visibility_hidden_from_you()}</p>
  {/if}
  {#if dataQueryable}
    <p class="note">{m.canvas_visibility_data_note()}</p>
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
