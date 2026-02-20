<script lang="ts">
  import APIIcon from "@rilldata/web-common/components/icons/APIIcon.svelte";
  import type { V1SecurityRule } from "@rilldata/web-common/runtime-client";

  export let id: string;
  export let title: string;
  export let description: string | undefined;
  export let resolver: string | undefined;
  export let openapiSummary: string | undefined;
  export let securityRules: V1SecurityRule[];
  export let reconcileError: string | undefined;

  $: accessExpression = securityRules?.[0]?.access?.conditionExpression;
</script>

<a
  href={`apis/${id}`}
  class="flex flex-col gap-y-1 group px-4 py-2.5 w-full h-full"
>
  <div class="flex gap-x-2 items-center min-h-[20px]">
    <APIIcon size="14px" />
    <span
      class="text-fg-primary text-sm font-semibold group-hover:text-accent-primary-action truncate"
    >
      {title}
    </span>
    {#if reconcileError}
      <span
        class="text-red-500 text-xs font-normal shrink-0"
        title={reconcileError}>Error</span
      >
    {/if}
  </div>
  <div
    class="flex gap-x-1 text-fg-secondary text-xs font-normal min-h-[16px] overflow-hidden"
  >
    {#if description}
      <span class="truncate">{description}</span>
      <span class="shrink-0">•</span>
    {:else if openapiSummary}
      <span class="truncate">{openapiSummary}</span>
      <span class="shrink-0">•</span>
    {/if}
    {#if resolver}
      <span class="shrink-0">{resolver}</span>
    {/if}
    {#if accessExpression}
      <span class="shrink-0">•</span>
      <span class="shrink-0 font-mono">access: {accessExpression}</span>
    {/if}
  </div>
</a>
