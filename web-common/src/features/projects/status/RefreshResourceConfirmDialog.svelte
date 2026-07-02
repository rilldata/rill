<script lang="ts">
  import {
    AlertDialog,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
  } from "@rilldata/web-common/components/alert-dialog/index.js";
  import { Button } from "@rilldata/web-common/components/button/index.js";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  export let open = false;
  export let name: string;
  export let onRefresh: () => Promise<void> | void;
  export let refreshType: "full" | "incremental" = "full";

  let isRefreshing = false;

  const MAX_NAME_LENGTH = 37;

  function truncateName(str: string): string {
    if (str.length > MAX_NAME_LENGTH) {
      return str.substring(0, MAX_NAME_LENGTH) + "...";
    }
    return str;
  }

  async function handleRefresh() {
    try {
      isRefreshing = true;
      await onRefresh();
      open = false;
    } catch (error) {
      console.error("Failed to refresh resource:", error);
    } finally {
      isRefreshing = false;
    }
  }
</script>

<AlertDialog bind:open>
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle>
        {refreshType === "full" ? m.status_action_full_refresh() : m.status_action_incremental_refresh()}
        <span class="font-semibold" title={name}>{truncateName(name)}</span>?
      </AlertDialogTitle>
      <AlertDialogDescription>
        <div class="mt-1">
          {#if refreshType === "full"}
            {m.status_full_refresh_warning()}
          {:else}
            {m.status_incremental_refresh_description()}
          {/if}
        </div>
      </AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter>
      <Button
        type="tertiary"
        disabled={isRefreshing}
        onClick={() => {
          open = false;
        }}>{m.status_cancel()}</Button
      >
      <Button
        type="primary"
        onClick={handleRefresh}
        disabled={isRefreshing}
        loading={isRefreshing}>{m.status_yes_refresh()}</Button
      >
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
