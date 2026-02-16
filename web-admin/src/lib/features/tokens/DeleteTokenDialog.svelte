<script lang="ts">
  import { Dialog } from "@rilldata/web-common/components/dialog-v2";
  import { Button } from "@rilldata/web-common/components/button";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";

  export let open = false;
  export let tokenName: string;
  export let tokenType: "service" | "user" = "service";
  export let onConfirm: () => Promise<void>;
  export let onClose: () => void = () => {
    open = false;
  };

  let loading = false;
  let error: string | null = null;

  async function handleConfirm() {
    loading = true;
    error = null;

    try {
      await onConfirm();
      loading = false;
      open = false;
      eventBus.emit("notification", {
        message: `${tokenType === "service" ? "Service" : "User"} token "${tokenName}" has been revoked.`,
      });
    } catch (e: unknown) {
      loading = false;
      if (e instanceof Error) {
        error = e.message;
      } else {
        error = "Failed to revoke token. Please try again.";
      }
    }
  }

  function handleCancel() {
    if (loading) return;
    error = null;
    onClose();
  }

  // Reset error state when dialog opens/closes
  $: if (!open) {
    error = null;
    loading = false;
  }
</script>

<Dialog bind:open on:close={handleCancel}>
  <div class="flex flex-col gap-4 p-6">
    <!-- Warning icon -->
    <div class="flex items-center gap-3">
      <div
        class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-full bg-red-100"
      >
        <svg
          class="h-5 w-5 text-red-600"
          xmlns="http://www.w3.org/2000/svg"
          viewBox="0 0 20 20"
          fill="currentColor"
          aria-hidden="true"
        >
          <path
            fill-rule="evenodd"
            d="M8.485 2.495c.673-1.167 2.357-1.167 3.03 0l6.28 10.875c.673 1.167-.168 2.625-1.516 2.625H3.72c-1.347 0-2.189-1.458-1.515-2.625L8.485 2.495zM10 6a.75.75 0 01.75.75v3.5a.75.75 0 01-1.5 0v-3.5A.75.75 0 0110 6zm0 9a1 1 0 100-2 1 1 0 000 2z"
            clip-rule="evenodd"
          />
        </svg>
      </div>
      <h2 class="text-lg font-semibold text-gray-900">Revoke Token</h2>
    </div>

    <!-- Warning message -->
    <p class="text-sm text-gray-600">
      Are you sure you want to revoke <strong class="text-gray-900"
        >{tokenName}</strong
      >? This action cannot be undone. Any integrations using this token will
      immediately lose access.
    </p>

    <!-- Error message -->
    {#if error}
      <div
        class="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700"
      >
        {error}
      </div>
    {/if}

    <!-- Action buttons -->
    <div class="flex justify-end gap-2 pt-2">
      <Button type="secondary" on:click={handleCancel} disabled={loading}>
        Cancel
      </Button>
      <Button type="destructive" on:click={handleConfirm} disabled={loading}>
        {#if loading}
          <svg
            class="mr-2 h-4 w-4 animate-spin"
            xmlns="http://www.w3.org/2000/svg"
            fill="none"
            viewBox="0 0 24 24"
          >
            <circle
              class="opacity-25"
              cx="12"
              cy="12"
              r="10"
              stroke="currentColor"
              stroke-width="4"
            />
            <path
              class="opacity-75"
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
            />
          </svg>
          Revoking…
        {:else}
          Revoke Token
        {/if}
      </Button>
    </div>
  </div>
</Dialog>