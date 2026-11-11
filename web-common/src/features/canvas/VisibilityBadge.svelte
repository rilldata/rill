<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { EyeOff } from "lucide-svelte";
  import type { Document } from "yaml";
  import { isHiddenPath, type YAMLPath } from "./layout-util";
  import type { CanvasEntity } from "./stores/canvas-entity";

  // Marks a canvas element that only some viewers see: a widget (in the corner of its card)
  // or a container (inline, in its chip).
  let {
    canvas,
    path,
    inline = false,
  }: {
    canvas: CanvasEntity;
    // The element's YAML path, e.g. ["rows", 0, "items", 1] or ["rows", 2].
    path: YAMLPath;
    inline?: boolean;
  } = $props();

  const parsedContent = $derived(canvas.parsedContent);
  const hiddenPaths = $derived(canvas.hiddenPaths);

  const conditions = $derived(collectConditions($parsedContent, path));
  const hiddenFromYou = $derived(isHiddenPath($hiddenPaths, path));

  // Collects the `if` conditions of the element and of its ancestors (tab group, tab,
  // container, row). Their YAML paths are the prefixes of the path that end at a list index.
  function collectConditions(
    doc: Document | undefined,
    path: YAMLPath,
  ): string[] {
    const conditions: string[] = [];
    for (let i = 2; i <= path.length; i += 2) {
      const raw: unknown = doc?.getIn([...path.slice(0, i), "if"]);
      if (typeof raw === "boolean") conditions.push(String(raw));
      else if (typeof raw === "string" && raw.trim() !== "")
        conditions.push(raw);
    }
    return conditions;
  }
</script>

{#if conditions.length > 0 || hiddenFromYou}
  <div
    class={["visibility-badge", { inline, "hidden-from-you": hiddenFromYou }]}
  >
    <Tooltip location="bottom" alignment="end" distance={4}>
      <EyeOff size={inline ? "12px" : "14px"} />
      <TooltipContent slot="tooltip-content" maxWidth="320px">
        {#if conditions.length > 0}
          <p>{m.canvas_visibility_badge()}</p>
          {#each conditions as condition, i (i)}
            <code class="condition">{condition}</code>
          {/each}
        {/if}
        {#if hiddenFromYou}
          <p>{m.canvas_visibility_hidden_from_you()}</p>
        {/if}
      </TooltipContent>
    </Tooltip>
  </div>
{/if}

<style lang="postcss">
  .visibility-badge {
    @apply absolute bottom-3.5 right-3.5 z-10 pointer-events-auto;
    @apply flex items-center justify-center size-6 rounded;
    @apply bg-surface-card border text-fg-secondary;
  }

  .visibility-badge.inline {
    @apply static size-5 border-0 bg-transparent;
  }

  .visibility-badge.hidden-from-you {
    @apply text-fg-primary border-fg-secondary;
  }

  .condition {
    @apply block font-mono text-xs break-all;
  }
</style>
