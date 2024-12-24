<!--
 To be used later when we have all the components in place.
 This dialog will be used as a quickstart option.
 -->

<script lang="ts">
  import * as AlertDialog from "@rilldata/web-common/components/alert-dialog";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { createResourceFile } from "../file-explorer/new-files";

  export let open = false;
  export let metricsViews: V1Resource[];
  export let wrapNavigation: (path: string | undefined) => Promise<void>;

  let selectedMetricsView: V1Resource | undefined = undefined;

  $: metricsViewOptions = metricsViews.map((resource) => ({
    value: resource.meta?.name?.name ?? "",
    label: resource.meta?.name?.name ?? "",
  }));

  async function createResource() {
    if (selectedMetricsView) {
      const newFilePath = await createResourceFile(
        ResourceKind.Canvas,
        selectedMetricsView,
      );
      await wrapNavigation(newFilePath);
    }
  }
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
      <AlertDialog.Cancel asChild let:builder>
        <Button large builders={[builder]} type="secondary">Cancel</Button>
      </AlertDialog.Cancel>

      <AlertDialog.Action asChild let:builder>
        <Button
          disabled={!selectedMetricsView}
          large
          builders={[builder]}
          type="primary"
          on:click={createResource}
        >
          Continue
        </Button>
      </AlertDialog.Action>
    </AlertDialog.Footer>
  </AlertDialog.Content>
</AlertDialog.Root>
