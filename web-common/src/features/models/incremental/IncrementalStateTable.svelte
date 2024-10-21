<script lang="ts">
  import type { V1Resource } from "../../../runtime-client";

  export let resource: V1Resource;

  $: schema = resource.model?.state?.incrementalStateSchema;
  $: state = resource.model?.state?.incrementalState;
</script>

{#if schema && schema.fields}
  <table>
    <thead>
      <tr>
        {#each schema.fields as field (field.name)}
          <th>{field.name}</th>
        {/each}
      </tr>
    </thead>
    {#if state}
      <tbody>
        <tr>
          {#each schema.fields as field (field.name)}
            <td>{field.name ? state[field.name] : "N/A"}</td>
          {/each}
        </tr>
      </tbody>
    {/if}
  </table>
{/if}
