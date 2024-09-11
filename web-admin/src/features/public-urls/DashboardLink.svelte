<script lang="ts">
  export let href: string;
  export let title: string;
</script>

{#if title}
  <a {href} class="text-blue-600 hover:underline">
    {title}
  </a>
{:else}
  <span>-</span>
{/if}
