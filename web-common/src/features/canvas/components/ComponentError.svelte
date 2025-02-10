<script lang="ts">
  import CrossIcon from "@rilldata/web-common/components/icons/CrossIcon.svelte";

  export let error: string | undefined;
</script>

<div class="component-error">
  <div class="error-icon">
    <CrossIcon size="20px" />
  </div>
  <div class="error-message">
    {error ?? "Component Error"}
  </div>
</div>

<style lang="postcss">
  .component-error {
    @apply flex flex-col gap-y-3 w-full h-full items-center justify-center p-4;
    @apply bg-red-50 text-red-600 rounded-lg border border-red-200;
  }

  .error-icon {
    @apply p-3 bg-red-100 rounded-full;
  }

  .error-message {
    @apply text-base font-medium text-center max-w-md;
  }
</style>
