<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { EyeOff } from "lucide-svelte";
  import type { Document } from "yaml";
  import type { BaseCanvasComponent } from "./components/BaseCanvasComponent";
  import { isHiddenPath } from "./layout-util";

  export let component: BaseCanvasComponent;

  $: ({ parsedContent, hiddenPaths } = component.parent);

  // The canvas item's YAML path: the component's path without its renderer key.
  $: itemPath = component.pathInYAML.slice(0, -1);

  $: conditions = collectConditions($parsedContent, itemPath);
  $: hiddenFromYou = isHiddenPath($hiddenPaths, itemPath);

  // Collects the `if` conditions of the item and of its ancestors (tab group, tab, row).
  // Their YAML paths are the prefixes of the item's path that end at a list index.
  function collectConditions(
    doc: Document | undefined,
    path: (string | number)[],
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
  <div class="visibility-badge" class:hidden-from-you={hiddenFromYou}>
    <Tooltip location="bottom" alignment="end" distance={4}>
      <EyeOff size="14px" />
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

  .visibility-badge.hidden-from-you {
    @apply text-fg-primary border-fg-secondary;
  }

  .condition {
    @apply block font-mono text-xs break-all;
  }
</style>
