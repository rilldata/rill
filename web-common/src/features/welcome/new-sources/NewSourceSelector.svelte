<script lang="ts">
  import { connectors } from "@rilldata/web-common/features/sources/modal/connector-schemas.ts";
  import { connectorIconMapping } from "@rilldata/web-common/features/connectors/connector-icon-mapping.ts";

  export let startConnectorSelection: (name: string) => void;
</script>

<div>
  <h2>Where is your data?</h2>
  <div class="grid grid-cols-3">
    {#each connectors as connector (connector.name)}
      {@const icon = connectorIconMapping[connector.name]}
      <button
        class="flex flex-row items-center gap-y-2"
        on:click={() => startConnectorSelection(connector.name)}
      >
        <svelte:component this={icon} size="24px" />
        <span class="text-sm">{connector.name}</span>
      </button>
    {/each}
  </div>
</div>
