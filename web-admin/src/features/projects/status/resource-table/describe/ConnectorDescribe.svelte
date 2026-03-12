<script lang="ts">
  import type { V1ConnectorV2 } from "@rilldata/web-common/runtime-client";
  import DescribeSection from "./DescribeSection.svelte";
  import DescribeRow from "./DescribeRow.svelte";

  export let connector: V1ConnectorV2;

  $: spec = connector?.spec;
  $: state = connector?.state;
  $: properties = spec?.properties ?? {};
  $: templatedProperties = new Set(spec?.templatedProperties ?? []);
  $: propertyKeys = Object.keys(properties).sort();

  // Separate env-var properties from regular ones
  $: envVarKeys = propertyKeys.filter((k) => templatedProperties.has(k));
  $: regularKeys = propertyKeys.filter((k) => !templatedProperties.has(k));
</script>

<div class="flex flex-col gap-y-3">
  {#if spec?.driver}
    <DescribeSection title="Connection">
      <DescribeRow label="Driver" value={spec.driver} />
    </DescribeSection>
  {/if}

  {#if envVarKeys.length > 0}
    <DescribeSection title="Environment Variables">
      {#each envVarKeys as key}
        <DescribeRow label={key} value={String(properties[key])} />
      {/each}
    </DescribeSection>
  {/if}

  {#if regularKeys.length > 0}
    <DescribeSection title="Properties">
      {#each regularKeys as key}
        <DescribeRow label={key} value={String(properties[key])} />
      {/each}
    </DescribeSection>
  {/if}

  {#if spec?.provisionArgs}
    <DescribeSection title="Provision Arguments">
      {#each Object.entries(spec.provisionArgs) as [key, val]}
        <DescribeRow label={key} value={String(val)} />
      {/each}
    </DescribeSection>
  {/if}

</div>
