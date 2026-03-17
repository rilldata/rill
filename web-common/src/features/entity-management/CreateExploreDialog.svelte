<script lang="ts">
  import * as AlertDialog from "@rilldata/web-common/components/alert-dialog";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { createResourceAndNavigate } from "../file-explorer/new-files";
  import { useRuntimeClient } from "../../runtime-client/v2";
  import { ResourceKind } from "./resource-selectors";

  export let open = false;
  export let metricsViews: V1Resource[];

  const runtimeClient = useRuntimeClient();

  let selectedMetricsView: V1Resource | undefined = undefined;

  $: metricsViewOptions = metricsViews.map((resource) => ({
    value: resource.meta?.name?.name ?? "",
    label: resource.meta?.name?.name ?? "",
  }));
</script>

<AlertDialog.Root bind:open>
  <AlertDialog.Content>
    <AlertDialog.Title>
      Which metrics view should this dashboard reference?
    </AlertDialog.Title>

    <AlertDialog.Description>
      This will determine the measures and dimensions you can explore on this
      dashboard.
    </AlertDialog.Description>

    <Select
      sameWidth
      options={metricsViewOptions}
      fontSize={14}
      placeholder="Select a metrics view"
      id="metrics-explore-selection"
      onChange={(value) => {
        selectedMetricsView = metricsViews.find(
          (resource) => resource.meta?.name?.name === value,
        );
      }}
    />

    <AlertDialog.Footer>
      <AlertDialog.Cancel>
        {#snippet child({ props })}
          <span style="display:contents" {...props}>
            <Button large type="secondary">Cancel</Button>
          </span>
        {/snippet}
      </AlertDialog.Cancel>

      <AlertDialog.Action>
        {#snippet child({ props })}
          <span style="display:contents" {...props}>
            <Button
              disabled={!selectedMetricsView}
              large
              type="primary"
              onClick={() =>
                void createResourceAndNavigate(
                  runtimeClient,
                  ResourceKind.Explore,
                  selectedMetricsView,
                )}
            >
              Continue
            </Button>
          </span>
        {/snippet}
      </AlertDialog.Action>
    </AlertDialog.Footer>
  </AlertDialog.Content>
</AlertDialog.Root>
