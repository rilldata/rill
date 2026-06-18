<script lang="ts">
  import ExternalLink from "@rilldata/web-common/components/icons/ExternalLink.svelte";

  export let value: string | null | undefined;
  export let href: string;
</script>

<div class="dimension-cell">
  <span class="truncate min-w-0">{value ?? "null"}</span>
  <a
    class="external-link"
    target="_blank"
    rel="noopener noreferrer"
    {href}
    title={href}
    onclick={(e) => e.stopPropagation()}
  >
    <ExternalLink className="fill-primary-600" />
  </a>
</div>

<style lang="postcss">
  .dimension-cell {
    @apply flex items-center gap-x-1 min-w-0;
  }

  .external-link {
    @apply inline-flex items-center shrink-0 opacity-0 transition-opacity;
  }

  .dimension-cell:hover .external-link {
    @apply opacity-100;
  }
</style>
