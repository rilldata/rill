<script lang="ts">
  import * as Alert from "@rilldata/web-common/components/alert-dialog/index.js";
  import Button from "web-common/src/components/button/Button.svelte";

  export let open = false;
  export let onChange: () => void;
</script>

<Alert.Root bind:open>
  <Alert.Trigger asChild>
    <div class="hidden"></div>
  </Alert.Trigger>
  <Alert.Content>
    <Alert.Header>
      <Alert.Title>Assign a new billing contact first</Alert.Title>
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
        on:click={() => {
          open = false;
        }}
      >
        Cancel
      </Button>
      <Button
        type="primary"
        on:click={() => {
          open = false;
          onChange();
        }}
      >
        Change billing contact
      </Button>
    </Alert.Footer>
  </Alert.Content>
</Alert.Root>
