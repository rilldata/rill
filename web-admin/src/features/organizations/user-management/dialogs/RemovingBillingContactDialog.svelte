<script lang="ts">
  import * as m from "@rilldata/web-common/paraglide/messages.js";
  import * as Alert from "@rilldata/web-common/components/alert-dialog";
  import Button from "web-common/src/components/button/Button.svelte";

  export let open = false;
  export let onChange: () => void;
</script>

<Alert.Root bind:open>
  <Alert.Trigger>
    {#snippet child({ props })}
      <div {...props} class="hidden"></div>
    {/snippet}
  </Alert.Trigger>
  <Alert.Content noCancel>
    <Alert.Header>
      <Alert.Title>
        {m.users_billing_remove_title()}
      </Alert.Title>
      <Alert.Description>
        <div class="mt-1">
          {m.users_billing_remove_desc()}
        </div>
      </Alert.Description>
    </Alert.Header>
    <Alert.Footer>
      <Button
        type="tertiary"
        onClick={() => {
          open = false;
        }}
      >
        {m.users_cancel()}
      </Button>
      <Button
        type="primary"
        onClick={() => {
          open = false;
          onChange();
        }}
      >
        {m.users_change_billing_contact()}
      </Button>
    </Alert.Footer>
  </Alert.Content>
</Alert.Root>
