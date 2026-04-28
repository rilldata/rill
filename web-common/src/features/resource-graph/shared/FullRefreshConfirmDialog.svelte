<script lang="ts">
  import * as AlertDialog from "@rilldata/web-common/components/alert-dialog";
  import { Button } from "@rilldata/web-common/components/button";

  export let open = false;
  export let resourceName: string;
  export let onConfirm: () => void;
</script>

<AlertDialog.Root bind:open>
  <AlertDialog.Content>
    <AlertDialog.Header>
      <AlertDialog.Title>Full Refresh {resourceName}?</AlertDialog.Title>
      <AlertDialog.Description>
        <div class="mt-1">
          A full refresh will re-ingest ALL data from scratch. This operation
          can take a significant amount of time and will update all dependent
          resources. Only proceed if you're certain this is necessary.
        </div>
      </AlertDialog.Description>
    </AlertDialog.Header>
    <AlertDialog.Footer>
      <Button
        type="secondary"
        onClick={() => {
          open = false;
        }}>Cancel</Button
      >
      <Button type="primary" onClick={onConfirm}>Yes, refresh</Button>
    </AlertDialog.Footer>
  </AlertDialog.Content>
</AlertDialog.Root>
