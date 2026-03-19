<!-- web-admin/src/features/admin/shared/StatusBadge.svelte -->
<script lang="ts">
  export let status: "success" | "error" | "pending" | "info" = "info";
  export let label: string;
</script>

<span class="badge" class:success={status === "success"}
  class:error={status === "error"} class:pending={status === "pending"}
  class:info={status === "info"}>
  {label}
</span>

<style lang="postcss">
  .badge {
    @apply inline-flex items-center px-2 py-0.5 text-xs font-medium rounded-full;
  }

  .success {
    @apply bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400;
  }

  .error {
    @apply bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400;
  }

  .pending {
    @apply bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400;
  }

  .info {
    @apply bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400;
  }
</style>
