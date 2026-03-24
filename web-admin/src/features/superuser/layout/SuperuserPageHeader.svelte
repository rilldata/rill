<script lang="ts">
  export let title: string;
  export let description: string = "";
</script>

<div class="mb-6">
  <h1 class="text-xl font-semibold text-fg-primary">
    {title}
  </h1>
  {#if description}
    <p class="text-sm text-fg-secondary mt-1">
      {description}
    </p>
  {/if}
</div>
