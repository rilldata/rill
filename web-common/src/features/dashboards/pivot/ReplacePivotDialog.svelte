<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import * as AlertDialog from "@rilldata/web-common/components/alert-dialog";

  export let open;
  export let onCancel: () => void;
  export let onReplace: () => void;
</script>

<AlertDialog.Root bind:open>
  <AlertDialog.Content>
    <AlertDialog.Title>Replace current pivot table?</AlertDialog.Title>

    <AlertDialog.Description>
      Starting a new table will loose your previous work. Bookmark tables you
      want to keep
    </AlertDialog.Description>

    <AlertDialog.Footer>
      <AlertDialog.Cancel asChild let:builder>
        <Button large builders={[builder]} type="secondary" on:click={onCancel}>
          Cancel
        </Button>
      </AlertDialog.Cancel>

      <AlertDialog.Action asChild let:builder>
        <Button large builders={[builder]} type="primary" on:click={onReplace}>
          Replace
        </Button>
      </AlertDialog.Action>
    </AlertDialog.Footer>
  </AlertDialog.Content>
</AlertDialog.Root>
