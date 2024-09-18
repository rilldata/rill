<script lang="ts">
  import {
    Button,
    IconSpaceFixer,
  } from "@rilldata/web-common/components/button";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import RefreshIcon from "@rilldata/web-common/components/icons/RefreshIcon.svelte";
  import ResponsiveButtonText from "@rilldata/web-common/components/panel/ResponsiveButtonText.svelte";
  import {
    V1ReconcileStatus,
    V1Resource,
    createRuntimeServiceCreateTrigger,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "../../../runtime-client/runtime-store";

  export let resource: V1Resource | undefined;
  export let collapse = false;

  const triggerMutation = createRuntimeServiceCreateTrigger();

  $: isIncrementalModel = resource?.model?.spec?.incremental;
  $: isModelIdle =
    resource?.meta?.reconcileStatus === V1ReconcileStatus.RECONCILE_STATUS_IDLE;

  function refreshModel(full: boolean) {
    void $triggerMutation.mutateAsync({
      instanceId: $runtime.instanceId,
      data: {
        models: [{ model: resource?.meta?.name?.name, full: full }],
      },
    });
  }
</script>

{#if isIncrementalModel}
  <DropdownMenu.Root>
    <DropdownMenu.Trigger asChild let:builder>
      <Button type="secondary" builders={[builder]} disabled={!isModelIdle}>
        <IconSpaceFixer pullLeft pullRight={collapse}>
          <RefreshIcon size="14px" />
        </IconSpaceFixer>
        <ResponsiveButtonText {collapse}>Refresh</ResponsiveButtonText>
        <IconSpaceFixer pullLeft pullRight={collapse}>
          <CaretDownIcon />
        </IconSpaceFixer>
      </Button>
    </DropdownMenu.Trigger>
    <DropdownMenu.Content>
      <DropdownMenu.Item on:click={() => refreshModel(false)}>
        Incremental refresh
      </DropdownMenu.Item>
      <DropdownMenu.Item on:click={() => refreshModel(true)}>
        Full refresh
      </DropdownMenu.Item>
    </DropdownMenu.Content>
  </DropdownMenu.Root>
{:else}
  <Button type="secondary" on:click={() => refreshModel(true)}>
    <IconSpaceFixer pullLeft pullRight={collapse}>
      <RefreshIcon size="14px" />
    </IconSpaceFixer>
    <ResponsiveButtonText {collapse}>Refresh</ResponsiveButtonText>
  </Button>
{/if}
