<script lang="ts">
  import * as AlertDialog from "@rilldata/web-common/components/alert-dialog";
  import { Button } from "@rilldata/web-common/components/button";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  export let open;
  export let onCancel: () => void;
  export let onReplace: () => void;
</script>

<AlertDialog.Root bind:open>
  <AlertDialog.Content>
    <AlertDialog.Title>{m.dashboard_replace_pivot_title()}</AlertDialog.Title>

    <AlertDialog.Description>
      {m.dashboard_replace_pivot_description()}
    </AlertDialog.Description>

    <AlertDialog.Footer>
      <AlertDialog.Cancel>
        {#snippet child({ props })}
          <Button {...props} large type="secondary" onClick={onCancel}
            >{m.dashboard_cancel()}</Button
          >
        {/snippet}
      </AlertDialog.Cancel>

      <AlertDialog.Action>
        {#snippet child({ props })}
          <Button {...props} large type="primary" onClick={onReplace}
            >{m.dashboard_replace()}</Button
          >
        {/snippet}
      </AlertDialog.Action>
    </AlertDialog.Footer>
  </AlertDialog.Content>
</AlertDialog.Root>
