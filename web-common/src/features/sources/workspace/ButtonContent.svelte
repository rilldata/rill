<script lang="ts">
  import IconSpaceFixer from "@rilldata/web-common/components/button/IconSpaceFixer.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import RefreshIcon from "@rilldata/web-common/components/icons/RefreshIcon.svelte";
  import ResponsiveButtonText from "@rilldata/web-common/components/panel/ResponsiveButtonText.svelte";
  import { quintOut } from "svelte/easing";
  import { slide } from "svelte/transition";

  export let collapse: boolean;
  export let hasUnsavedChanges: boolean;
  export let isLocalFileConnector: boolean;
</script>

<IconSpaceFixer pullLeft pullRight={collapse}>
  <RefreshIcon size="14px" />
</IconSpaceFixer>
<ResponsiveButtonText {collapse}>
  <div class="flex">
    {#if hasUnsavedChanges}
      <span
        class="pr-1 w-fit whitespace-nowrap"
        transition:slide={{
          duration: 250,
          axis: "x",
          easing: quintOut,
        }}
      >
        Save and
      </span>
    {/if}
    <span class:lowercase={hasUnsavedChanges}>Refresh</span>
  </div>
</ResponsiveButtonText>
{#if !hasUnsavedChanges && isLocalFileConnector}
  <CaretDownIcon size="14px" />
{/if}
