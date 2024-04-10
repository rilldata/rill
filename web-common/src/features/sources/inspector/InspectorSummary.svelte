<script lang="ts">
  import CollapsibleSectionTitle from "@rilldata/web-common/layout/CollapsibleSectionTitle.svelte";

  export let rowCount: string;
  export let columnCount: string;
  export let showTitle: boolean = true;

  let showModelInformation = true;
</script>

<div class="size-full px-4">
  {#if showTitle}
    <CollapsibleSectionTitle
      tooltipText="model summary"
      bind:active={showModelInformation}
    >
      Model summary
    </CollapsibleSectionTitle>
  {/if}
  {#if showModelInformation}
    <table class="size-full text-gray-800">
      <tbody>
        <tr>
          <td>
            <slot name="row-header" />
          </td>
          <td class="text-right">
            <b>{rowCount}</b>
          </td>
        </tr>
        <tr>
          <td>
            <slot name="column-header" />
          </td>
          <td class="text-right">
            <b>{columnCount}</b>
          </td>
        </tr>
      </tbody>
    </table>
  {/if}
</div>
