<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import { createEventDispatcher } from "svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import RefreshIcon from "@rilldata/web-common/components/icons/RefreshIcon.svelte";
  import { allowPrimary } from "../../dashboards/workspace/DeployProjectCTA.svelte";
  import { useModels } from "../../models/selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { removeLeadingSlash } from "../../entity-management/entity-mappers";
  import {
    resourceColorMapping,
    resourceIconMapping,
  } from "../../entity-management/resource-icon-mapping";
  import Add from "@rilldata/web-common/components/icons/Add.svelte";

  const dispatch = createEventDispatcher();

  export let hasErrors: boolean;
  export let hasUnsavedChanges: boolean;
  export let isLocalFileConnector: boolean;
  export let sourceName: string;

  $: ({ instanceId } = $runtime);

  $: modelsQuery = useModels(instanceId);

  $: modelsForSource = ($modelsQuery.data ?? []).filter((model) =>
    model.meta?.refs?.some((ref) => ref.name === sourceName),
  );
</script>

{#if !isLocalFileConnector || hasUnsavedChanges}
  <Tooltip distance={8}>
    <Button
      square
      on:click={() => {
        if (isLocalFileConnector && !hasUnsavedChanges) return;
        if (hasUnsavedChanges) {
          dispatch("save-source");
        } else {
          dispatch("refresh-source");
        }
      }}
      label="Refresh"
      type="secondary"
      disabled={hasUnsavedChanges}
    >
      <RefreshIcon size="14px" />
    </Button>

    <TooltipContent slot="tooltip-content">
      {#if hasUnsavedChanges}
        Save your changes to refresh
      {:else}
        Refresh source
      {/if}
    </TooltipContent>
  </Tooltip>
{:else}
  <DropdownMenu.Root>
    <DropdownMenu.Trigger asChild let:builder>
      <Tooltip distance={8}>
        <Button
          builders={[builder]}
          label="Refresh"
          type={hasUnsavedChanges ? "primary" : "secondary"}
        >
          <RefreshIcon size="14px" />
        </Button>
        <TooltipContent slot="tooltip-content">Refresh source</TooltipContent>
      </Tooltip>
    </DropdownMenu.Trigger>

    <DropdownMenu.Content>
      <DropdownMenu.Item
        on:click={() => {
          dispatch("refresh-source");
        }}
      >
        Refresh source
      </DropdownMenu.Item>
      <DropdownMenu.Item on:click={() => dispatch("replace-source")}>
        Replace source with uploaded file
      </DropdownMenu.Item>
    </DropdownMenu.Content>
  </DropdownMenu.Root>
{/if}

{#if modelsForSource.length === 0}
  <Button
    disabled={hasUnsavedChanges || hasErrors}
    on:click={() => dispatch("create-model")}
    type={$allowPrimary ? "primary" : "secondary"}
  >
    Create model
  </Button>
{:else}
  <DropdownMenu.Root>
    <DropdownMenu.Trigger asChild let:builder>
      <Button builders={[builder]} type="secondary">
        Go to
        <CaretDownIcon />
      </Button>
    </DropdownMenu.Trigger>

    <DropdownMenu.Content align="end">
      {#each modelsForSource as resource (resource?.meta?.name?.name)}
        {@const filePath = resource?.meta?.filePaths?.[0]}
        {@const resourceKind = resource?.meta?.name?.kind}
        {#if filePath && resourceKind}
          <DropdownMenu.Item href={`/files/${removeLeadingSlash(filePath)}`}>
            <svelte:component
              this={resourceIconMapping[resourceKind]}
              color={resourceColorMapping[resourceKind]}
              size="14px"
            />
            {resource?.meta?.name?.name ?? "Loading..."}
          </DropdownMenu.Item>
        {/if}
      {/each}
      <DropdownMenu.Separator />
      <DropdownMenu.Item on:click={() => dispatch("create-model")}>
        <Add />
        Create model
      </DropdownMenu.Item>
    </DropdownMenu.Content>
  </DropdownMenu.Root>
{/if}
