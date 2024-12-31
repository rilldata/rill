<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Alert from "@rilldata/web-common/components/icons/Alert.svelte";

  export let saving: boolean;
  export let errorMessage: string | undefined;
  export let onAcceptIncoming: () => void;
  export let onAcceptCurrent: () => void;
</script>

<header class="flex w-full border-b">
  <div class="border-r">
    <h2 class="italic text-gray-400">Unsaved changes</h2>

    <Button
      type="subtle"
      loading={saving}
      loadingCopy="Saving"
      danger={!!errorMessage && !saving}
      disabled={saving}
      on:click={onAcceptCurrent}
    >
      {#if errorMessage}
        <Alert size="14px" />
        {errorMessage} Try again.
      {:else}
        Accept current
      {/if}
    </Button>
  </div>

  <div>
    <h2>Incoming content</h2>
    <Button type="primary" on:click={onAcceptIncoming}>Accept incoming</Button>
  </div>
</header>

<style lang="postcss">
  h2 {
    @apply text-sm font-semibold;
  }

  div {
    @apply w-full p-1.5 pl-3 flex justify-between items-center;
  }
</style>
