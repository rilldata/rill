<script lang="ts">
  import AlertTriangle from "@rilldata/web-common/components/icons/AlertTriangle.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import TooltipTitle from "@rilldata/web-common/components/tooltip/TooltipTitle.svelte";

  export let modelHasError = false;
</script>

<Tooltip location="left" alignment="start" distance={16}>
  <slot />
  <TooltipContent slot="tooltip-content" maxWidth="300px">
    <TooltipTitle>
      <svelte:fragment slot="name"
        ><slot name="tooltip-title" /></svelte:fragment
      >
      <svelte:fragment slot="description">
        <slot name="tooltip-right" />
      </svelte:fragment>
    </TooltipTitle>
    <div class="pb-1 leading-4">
      <p class="text-gray-200">
        <slot name="tooltip-description" />
      </p>
      {#if modelHasError}
        <p class="italic pt-2 text-gray-100">
          <span
            class="inline-grid place-items-center"
            style:transform="translateY(1px)"
            ><AlertTriangle size="12px" /></span
          > The model has an error. Showing the last valid model results.
        </p>
      {/if}
    </div>
  </TooltipContent>
</Tooltip>
