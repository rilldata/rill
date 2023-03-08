<script lang="ts">
  import { goto } from "$app/navigation";
  import {
    Button,
    IconSpaceFixer,
  } from "@rilldata/web-common/components/button";
  import { WithTogglableFloatingElement } from "@rilldata/web-common/components/floating-element";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import Forward from "@rilldata/web-common/components/icons/Forward.svelte";
  import { Menu, MenuItem } from "@rilldata/web-common/components/menu";
  import { RuntimeUrl } from "@rilldata/web-local/lib/application-state-stores/initialize-node-store-contexts";

  import ResponsiveButtonText from "@rilldata/web-common/components/panel/ResponsiveButtonText.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { runtime } from "../../../runtime-client/runtime-store";
  import CreateDashboardButton from "./CreateDashboardButton.svelte";

  export let availableDashboards;
  export let modelName: string;
  export let suppressTooltips = false;
  export let modelHasError = false;

  export let collapse = false;

  const onExport = async (exportExtension: "csv" | "parquet") => {
    // TODO: how do we handle errors ?
    window.open(
      `${RuntimeUrl}/v1/instances/${$runtime.instanceId}/table/${modelName}/export/${exportExtension}`
    );
  };
</script>

<Tooltip
  alignment="middle"
  distance={16}
  location="left"
  suppress={suppressTooltips}
>
  <!-- attach floating element right here-->
  <WithTogglableFloatingElement
    alignment="end"
    bind:active={suppressTooltips}
    distance={8}
    let:toggleFloatingElement
    location="bottom"
  >
    <Button
      disabled={modelHasError}
      on:click={toggleFloatingElement}
      type="secondary"
    >
      <IconSpaceFixer pullLeft pullRight={collapse}
        ><CaretDownIcon /></IconSpaceFixer
      >

      <ResponsiveButtonText {collapse}>Export</ResponsiveButtonText>
    </Button>
    <Menu
      dark
      on:click-outside={toggleFloatingElement}
      on:escape={toggleFloatingElement}
      slot="floating-element"
    >
      <MenuItem
        on:select={() => {
          toggleFloatingElement();
          onExport("parquet");
        }}
      >
        Export as Parquet
      </MenuItem>
      <MenuItem
        on:select={() => {
          toggleFloatingElement();
          onExport("csv");
        }}
      >
        Export as CSV
      </MenuItem>
    </Menu>
  </WithTogglableFloatingElement>
  <TooltipContent slot="tooltip-content">
    {#if modelHasError}Fix the errors in your model to export
    {:else}
      Export the modeled data as a file
    {/if}
  </TooltipContent>
</Tooltip>

{#if availableDashboards?.length === 0}
  <CreateDashboardButton {collapse} hasError={modelHasError} {modelName} />
{:else if availableDashboards?.length === 1}
  <Tooltip distance={8} alignment="end">
    <Button
      on:click={() => {
        goto(`/dashboard/${availableDashboards[0].name}`);
      }}
    >
      <IconSpaceFixer pullLeft pullRight={collapse}>
        <Forward />
      </IconSpaceFixer>
      <ResponsiveButtonText {collapse}>Go to Dashboard</ResponsiveButtonText>
    </Button>
    <TooltipContent slot="tooltip-content">
      Go to the dashboard associated with this model
    </TooltipContent>
  </Tooltip>
{:else}
  <Tooltip distance={8} alignment="end">
    <WithTogglableFloatingElement
      let:toggleFloatingElement
      distance={8}
      alignment="end"
    >
      <Button on:click={toggleFloatingElement}>
        <IconSpaceFixer pullLeft pullRight={collapse}>
          <Forward /></IconSpaceFixer
        >
        <ResponsiveButtonText {collapse}>Go to Dashboard</ResponsiveButtonText>
      </Button>
      <Menu
        dark
        slot="floating-element"
        on:escape={toggleFloatingElement}
        on:click-outside={toggleFloatingElement}
      >
        {#each availableDashboards as dashboard}
          <MenuItem
            on:select={() => {
              goto(`/dashboard/${dashboard.name}`);
              toggleFloatingElement();
            }}
          >
            {dashboard.name}
          </MenuItem>
        {/each}
      </Menu>
    </WithTogglableFloatingElement>
    <TooltipContent slot="tooltip-content">
      Go to one of {availableDashboards.length} dashboards associated with this model
    </TooltipContent>
  </Tooltip>
{/if}
