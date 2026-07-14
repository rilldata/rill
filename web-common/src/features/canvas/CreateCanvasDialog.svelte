<!--
 To be used later when we have all the components in place.
 This dialog will be used as a quickstart option.
 -->

<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import * as AlertDialog from "@rilldata/web-common/components/alert-dialog";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "../../runtime-client/v2";
  import { createResourceFile } from "../entity-management/add/new-files.ts";

  export let open = false;
  export let metricsViews: V1Resource[];
  export let wrapNavigation: (path: string | undefined) => Promise<void>;

  const runtimeClient = useRuntimeClient();

  let selectedMetricsView: V1Resource | undefined = undefined;

  $: metricsViewOptions = metricsViews.map((resource) => ({
    value: resource.meta?.name?.name ?? "",
    label: resource.meta?.name?.name ?? "",
  }));

  async function createResource() {
    if (selectedMetricsView) {
      const newFilePath = await createResourceFile(
        runtimeClient,
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
      {m.canvas_which_metrics_view()}
    </AlertDialog.Title>

    <AlertDialog.Description>
      {m.canvas_metrics_view_description()}
    </AlertDialog.Description>

    <Select
      sameWidth
      options={metricsViewOptions}
      fontSize={14}
      placeholder={m.canvas_select_metrics_view()}
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
          <Button {...props} large type="secondary">{m.common_cancel()}</Button>
        {/snippet}
      </AlertDialog.Cancel>

      <AlertDialog.Action>
        {#snippet child({ props })}
          <Button
            {...props}
            disabled={!selectedMetricsView}
            large
            type="primary"
            onClick={createResource}
          >
            {m.common_continue()}
          </Button>
        {/snippet}
      </AlertDialog.Action>
    </AlertDialog.Footer>
  </AlertDialog.Content>
</AlertDialog.Root>
