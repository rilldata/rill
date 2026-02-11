<script lang="ts">
  import type { ComponentType } from "svelte";

  export let icon: ComponentType;

  export let message: string;
  export let action: string = "";
</script>

<div class="flex flex-col items-center gap-y-4">
  <svelte:component this={icon} size="32px" />
  <div class="flex flex-col gap-y-2 items-center text-sm">
    <div class="text-fg-secondary font-semibold">
      {message}
    </div>
    {#if $$slots.action}
      <div class="text-fg-muted">
        <slot name="action" />
      </div>
    {:else if action}
      <div class="text-fg-muted">
        {action}
      </div>
    {/if}
  </div>
</div>
