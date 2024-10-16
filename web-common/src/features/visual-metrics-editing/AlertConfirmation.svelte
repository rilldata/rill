<script lang="ts">
  import * as AlertDialog from "@rilldata/web-common/components/alert-dialog";
  import type { Confirmation } from "./lib";
  import Button from "@rilldata/web-common/components/button/Button.svelte";

  export let confirmation: Confirmation;
  export let onCancel: () => void;
  export let onConfirm: () => void;
</script>

<AlertDialog.Root open>
  <AlertDialog.Content>
    <AlertDialog.Header>
      <AlertDialog.Title>
        {#if confirmation.action === "delete"}
          <h2>
            Delete {confirmation.index === undefined
              ? "selected items"
              : "this " + confirmation.type?.slice(0, -1)}?
          </h2>
        {:else if confirmation.action === "cancel"}
          <h2>Cancel changes to {confirmation.type?.slice(0, -1)}?</h2>
        {:else if confirmation.action === "switch"}
          <h2>Switch reference model?</h2>
        {/if}
      </AlertDialog.Title>
      <AlertDialog.Description>
        {#if confirmation.action === "cancel"}
          You haven't saved changes to this {confirmation.type?.slice(0, -1)} yet,
          so closing this window will lose your work.
        {:else if confirmation.action === "delete"}
          You will permanently remove {confirmation.index === undefined
            ? "the selected items"
            : "this " + confirmation.type?.slice(0, -1)} from all associated dashboards.
        {:else if confirmation.action === "switch"}
          Switching to a different model may break your measures and dimensions
          unless the new model has similar data.
        {/if}
      </AlertDialog.Description>
    </AlertDialog.Header>
    <AlertDialog.Footer class="gap-y-2">
      <AlertDialog.Cancel asChild let:builder>
        <Button
          builders={[builder]}
          type="secondary"
          large
          gray={confirmation.action === "delete"}
          on:click={onCancel}
        >
          {#if confirmation.action === "cancel"}Keep editing{:else}Cancel{/if}
        </Button>
      </AlertDialog.Cancel>

      <AlertDialog.Action asChild let:builder>
        <Button
          large
          builders={[builder]}
          status={confirmation.action === "delete" ? "error" : "info"}
          type="primary"
          on:click={onConfirm}
        >
          {#if confirmation.action === "delete"}
            Yes, delete
          {:else if confirmation.action === "switch"}
            Switch model
          {:else if confirmation.action === "cancel" && confirmation.field}
            Switch items
          {:else}
            Close
          {/if}
        </Button>
      </AlertDialog.Action>
    </AlertDialog.Footer>
  </AlertDialog.Content>
</AlertDialog.Root>

<style lang="postcss">
  h2 {
    @apply font-semibold text-lg;
  }
</style>
