<script lang="ts">
  import * as Alert from "@rilldata/web-common/components/alert-dialog";
  import Button from "web-common/src/components/button/Button.svelte";

  export let open = false;
  export let onChange: () => void;
</script>

<Alert.Root bind:open>
  <Alert.Trigger asChild>
    <div class="hidden"></div>
  </Alert.Trigger>
  <Alert.Content noCancel>
    <Alert.Header>
      <Alert.Title>
        Assign a new billing contact first to remove this user
      </Alert.Title>
      <Alert.Description>
        <div class="mt-1">
          This user is the current billing contact and can’t be removed. To
          proceed, assign another admin as the billing contact.
        </div>
      </Alert.Description>
    </Alert.Header>
    <Alert.Footer>
      <Button
        type="plain"
        onClick={() => {
          open = false;
        }}
      >
        Cancel
      </Button>
      <Button
        type="primary"
        onClick={() => {
          open = false;
          onChange();
        }}
      >
        Change billing contact
      </Button>
    </Alert.Footer>
  </Alert.Content>
</Alert.Root>
