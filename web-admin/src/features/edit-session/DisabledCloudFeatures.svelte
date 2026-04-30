<script lang="ts" context="module">
  import type { ComponentType } from "svelte";

  export type CloudFeature = {
    label: string;
    icon?: ComponentType;
    compact?: boolean;
    square?: boolean;
  };
</script>

<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";

  export let features: CloudFeature[];
  export let cta: string;
</script>

{#each features as { label, icon, compact, square } (label)}
  <Tooltip distance={8}>
    <Button type="secondary" {compact} {square} disabled {label}>
      {#if icon}
        <svelte:component this={icon} size="16px" class="flex-none" />
      {:else}
        {label}
      {/if}
    </Button>
    <TooltipContent slot="tooltip-content" maxWidth="240px">
      <span class="text-xs">{cta}</span>
    </TooltipContent>
  </Tooltip>
{/each}
