<script lang="ts">
  import LoadingCircleOutline from "@rilldata/web-common/components/icons/LoadingCircleOutline.svelte";

  export let onclick: () => void;
  export let loading: boolean;
  export let disabled: boolean;
  export let label: string;
</script>

<button {onclick} {disabled} aria-label={label}>
  {#if loading}
    <LoadingCircleOutline size="16px" color="var(--color-primary-600)" />
  {:else}
    <slot name="icon" />
  {/if}
  <slot />
</button>

<style lang="postcss">
  button {
    @apply flex flex-row items-center justify-center gap-2 px-4 py-2 h-fit;
    @apply text-sm bg-surface-overlay rounded-md border text-fg-secondary;
  }
  button:hover {
    @apply border-accent-primary-action shadow-lg cursor-pointer;
  }
</style>
