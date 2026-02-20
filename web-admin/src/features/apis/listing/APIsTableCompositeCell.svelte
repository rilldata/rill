<script lang="ts">
  import APIIcon from "@rilldata/web-common/components/icons/APIIcon.svelte";

  export let id: string;
  export let title: string;
  export let resolver: string | undefined;
  export let openapiSummary: string | undefined;
  export let securityRuleCount: number;
  export let reconcileError: string | undefined;
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
    <!-- TODO: add status icon (check/error) once API execution state is tracked -->
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
    {#if resolver}
      <span class="shrink-0">{resolver}</span>
    {/if}
    {#if openapiSummary}
      <span class="shrink-0">•</span>
      <span class="truncate">{openapiSummary}</span>
    {/if}
    {#if securityRuleCount > 0}
      <span class="shrink-0">•</span>
      <span class="shrink-0"
        >{securityRuleCount} security rule{securityRuleCount > 1
          ? "s"
          : ""}</span
      >
    {/if}
    <!-- TODO: add last invocation time and owner info -->
  </div>
</a>
