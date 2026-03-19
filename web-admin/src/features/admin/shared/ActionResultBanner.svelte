<!-- web-admin/src/features/admin/shared/ActionResultBanner.svelte -->
<script lang="ts">
  export let type: "success" | "error" | "" = "";
  export let message: string = "";

  export function show(newType: "success" | "error", newMessage: string) {
    type = newType;
    message = newMessage;
    setTimeout(() => {
      type = "";
      message = "";
    }, 5000);
  }
</script>

{#if type && message}
  <div
    class="banner"
    class:success={type === "success"}
    class:error={type === "error"}
  >
    <span>{message}</span>
    <button
      class="close-btn"
      on:click={() => {
        type = "";
        message = "";
      }}>x</button
    >
  </div>
{/if}

<style lang="postcss">
  .banner {
    @apply flex items-center justify-between px-4 py-3 rounded-md text-sm mb-4;
  }

  .success {
    @apply bg-green-50 text-green-800 dark:bg-green-900/20 dark:text-green-300;
  }

  .error {
    @apply bg-red-50 text-red-800 dark:bg-red-900/20 dark:text-red-300;
  }

  .close-btn {
    @apply ml-4 text-current opacity-50 hover:opacity-100;
  }
</style>
