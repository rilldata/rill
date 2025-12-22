<script lang="ts">
  import type { ComponentType } from "svelte";

  export let icon: ComponentType;
  export let iconColor: string;
  export let message: string;
  export let action: string = "";
</script>

<div class="flex flex-col items-center gap-y-4">
  <svelte:component this={icon} size="32px" color={iconColor} />
  <div class="flex flex-col gap-y-2 items-center text-sm">
    <div class="text-gray-600 font-semibold">
      {message}
    </div>
    {#if $$slots.action}
      <div class="text-gray-500">
        <slot name="action" />
      </div>
    {:else if action}
      <div class="text-gray-500">
        {action}
      </div>
    {/if}
  </div>
</div>
