<script lang="ts">
  import * as AlertDialog from "@rilldata/web-common/components/alert-dialog";
  import { Button } from "@rilldata/web-common/components/button";

  export let open;
  export let onCancel: () => void;
  export let onReplace: () => void;
</script>

<AlertDialog.Root bind:open>
  <AlertDialog.Content>
    <AlertDialog.Title>Replace current pivot table?</AlertDialog.Title>

    <AlertDialog.Description>
      Starting a new table will lose your previous work. Bookmark tables you
      want to keep
    </AlertDialog.Description>

    <AlertDialog.Footer>
      <AlertDialog.Cancel asChild let:builder>
        <Button large builders={[builder]} type="secondary" onClick={onCancel}>
          Cancel
        </Button>
      </AlertDialog.Cancel>

      <AlertDialog.Action asChild let:builder>
        <Button large builders={[builder]} type="primary" onClick={onReplace}>
          Replace
        </Button>
      </AlertDialog.Action>
    </AlertDialog.Footer>
  </AlertDialog.Content>
</AlertDialog.Root>
