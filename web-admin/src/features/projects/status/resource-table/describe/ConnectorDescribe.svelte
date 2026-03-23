<script lang="ts">
  import type { V1ConnectorV2 } from "@rilldata/web-common/runtime-client";
  import DescribeSection from "./DescribeSection.svelte";
  import DescribeRow from "./DescribeRow.svelte";
  import { formatPropertyValue } from "./utils";

  export let connector: V1ConnectorV2;

  $: spec = connector?.spec;
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
      {#each envVarKeys as key (key)}
        <DescribeRow label={key} value={formatPropertyValue(properties[key])} />
      {/each}
    </DescribeSection>
  {/if}

  {#if regularKeys.length > 0}
    <DescribeSection title="Properties">
      {#each regularKeys as key (key)}
        <DescribeRow label={key} value={formatPropertyValue(properties[key])} />
      {/each}
    </DescribeSection>
  {/if}

  {#if spec?.provisionArgs}
    <DescribeSection title="Provision Arguments">
      {#each Object.entries(spec.provisionArgs) as [key, val] (key)}
        <DescribeRow label={key} value={formatPropertyValue(val)} />
      {/each}
    </DescribeSection>
  {/if}
</div>
