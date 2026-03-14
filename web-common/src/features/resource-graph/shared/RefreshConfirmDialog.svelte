<script lang="ts">
  import * as AlertDialog from "@rilldata/web-common/components/alert-dialog";
  import Button from "@rilldata/web-common/components/button/Button.svelte";

  export let open = false;
  export let onRefresh: () => void;
</script>

<AlertDialog.Root bind:open>
  <AlertDialog.Content>
    <AlertDialog.Header>
      <AlertDialog.Title>Refresh all sources and models?</AlertDialog.Title>
      <AlertDialog.Description>
        <div class="flex flex-col gap-y-2 mt-1">
          <p>This will refresh all project sources and models.</p>
          <p>
            <span class="font-medium">Note:</span> To refresh a single resource,
            click the '...' button on a node and select the refresh option.
          </p>
        </div>
      </AlertDialog.Description>
    </AlertDialog.Header>
    <AlertDialog.Footer>
      <Button
        type="tertiary"
        onClick={() => {
          open = false;
        }}>Cancel</Button
      >
      <Button
        type="primary"
        onClick={() => {
          onRefresh();
          open = false;
        }}>Yes, refresh</Button
      >
    </AlertDialog.Footer>
  </AlertDialog.Content>
</AlertDialog.Root>
