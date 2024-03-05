<script lang="ts">
  export let title: string;
  export let description: string = "";
  export let padding = "p-3";
</script>

<div class="flex flex-col bg-white {padding} gap-y-3 rounded">
  <div class="flex flex-col">
    <span class="text-base text-medium text-slate-900">{title}</span>
    {#if description}
      <span class="text-sm text-slate-600">{description}</span>
    {/if}
  </div>
  <slot />
</div>
