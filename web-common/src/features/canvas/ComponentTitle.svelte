<script lang="ts">
  export let title: string | undefined = undefined;
  export let description: string | undefined = undefined;

  export let faint: boolean = false;
</script>

{#if title || description}
  <div class="w-full h-fit flex flex-col bg-white p-2">
    {#if title}
      <h1 class:faint class="title">{title}</h1>
    {/if}
    {#if description}
      <h2 class="description">{description}</h2>
    {/if}
  </div>
{/if}

<style lang="postcss">
  h1 {
    font-size: 16px;
    font-weight: 500;
  }

  h2 {
    font-size: 12px;
    font-weight: 400;
  }
  .title {
    font-size: 15px;
    @apply font-medium text-gray-800;
  }

  .title.faint {
    @apply text-gray-500;
  }

  .description {
    font-size: 13px;
    @apply text-gray-600 leading-none pt-0.5;
  }
</style>
