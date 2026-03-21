<script lang="ts">
  export let title: string;
  export let description: string = "";
</script>

<div class="mb-6">
  <h1 class="text-xl font-semibold text-slate-900 dark:text-slate-100">
    {title}
  </h1>
  {#if description}
    <p class="text-sm text-slate-500 dark:text-slate-400 mt-1">
      {description}
    </p>
  {/if}
</div>
