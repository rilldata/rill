<script lang="ts">
  export let href: string;
  export let title: string;
</script>

{#if title}
  <a
    {href}
    class="text-primary-600 hover:underline"
    target="_blank"
    rel="noopener noreferrer"
  >
    {title}
  </a>
{:else}
  <span>-</span>
{/if}
