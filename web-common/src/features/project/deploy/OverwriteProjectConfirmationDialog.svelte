<script lang="ts">
  import * as Alert from "@rilldata/web-common/components/alert-dialog/index.js";
  import { Button } from "@rilldata/web-common/components/button";

  export let open: boolean;
  export let rillManagedProject: boolean;
  export let deployUrl: string;
</script>

<Alert.Root bind:open>
  <Alert.Trigger asChild>
    <div class="hidden"></div>
  </Alert.Trigger>
  <Alert.Content class="min-w-[600px]">
    <Alert.Header>
      <Alert.Title>
        Are you sure you want to overwrite this project?
      </Alert.Title>
      <Alert.Description>
        {#if rillManagedProject}
          Existing project files will be replaced with new ones and cannot be
          retrieved again.
        {:else}
          This action will disconnect the existing Github-managed project and
          replace it.
        {/if}
      </Alert.Description>
    </Alert.Header>
    <Alert.Footer class="mt-5">
      <Button onClick={() => (open = false)} type="secondary">Cancel</Button>
      <Button
        type="primary"
        status={rillManagedProject ? "error" : "info"}
        href={deployUrl}
      >
        Yes, overwrite
      </Button>
    </Alert.Footer>
  </Alert.Content>
</Alert.Root>
