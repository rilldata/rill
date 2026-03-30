<script lang="ts">
  import * as Alert from "@rilldata/web-common/components/alert-dialog";
  import Button from "../../../components/button/Button.svelte";
  import type {
    CreateModelStep,
    ExploreConnectorStep,
  } from "@rilldata/web-common/features/add-data/manager/steps/types.ts";
  import { connectorInfoMap } from "@rilldata/web-common/features/sources/modal/connector-schemas.ts";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { maybeDeleteConnector } from "@rilldata/web-common/features/add-data/manager/steps/connector.ts";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";

  let {
    open = $bindable(),
    stepState,
    onClose,
  }: {
    open: boolean;
    stepState: CreateModelStep | ExploreConnectorStep;
    onClose: () => void;
  } = $props();
  let connector = $derived(stepState.connector);
  let schema = $derived(stepState.schema);
  let schemaInfo = $derived(connectorInfoMap.get(schema));

  const runtimeClient = useRuntimeClient();

  async function handleRemoveConnector() {
    await maybeDeleteConnector(runtimeClient, queryClient, connector);
    onClose();
  }
</script>

<Alert.Root bind:open>
  <Alert.Trigger>
    {#snippet child({ props })}
      <div {...props} class="hidden"></div>
    {/snippet}
  </Alert.Trigger>
  <Alert.Content>
    <Alert.Title>Delete created connector?</Alert.Title>
    <Alert.Description>
      A {schemaInfo?.displayName ?? schema} connector {connector} was created, do
      you want to remove it or keep before starting over?
    </Alert.Description>
    <Alert.Footer>
      <Button type="tertiary" onClick={onClose}>Keep it</Button>
      <Button type="primary" onClick={() => void handleRemoveConnector()}>
        Remove it
      </Button>
    </Alert.Footer>
  </Alert.Content>
</Alert.Root>
