<script lang="ts">
  import type { V1Component } from "@rilldata/web-common/runtime-client";
  import DescribeSection from "./DescribeSection.svelte";
  import DescribeRow from "./DescribeRow.svelte";

  export let component: V1Component;

  $: spec = component?.spec;
  $: state = component?.state;
  $: inputs = spec?.input ?? [];
  $: rendererProps = spec?.rendererProperties ?? {};
  $: rendererPropKeys = Object.keys(rendererProps).sort();
</script>

<div class="flex flex-col gap-y-3">
  <!-- General -->
  <DescribeSection title="General">
    <DescribeRow label="Renderer" value={spec?.renderer} />
    <DescribeRow
      label="Defined in canvas"
      value={String(!!spec?.definedInCanvas)}
    />
    {#if state?.dataRefreshedOn}
      <DescribeRow
        label="Data refreshed on"
        value={new Date(state.dataRefreshedOn).toLocaleString()}
        mono={false}
      />
    {/if}
  </DescribeSection>

  <!-- Renderer Properties -->
  <DescribeSection title="Renderer Properties">
    {#if rendererPropKeys.length > 0}
      {#each rendererPropKeys as key}
        {@const val = rendererProps[key]}
        <DescribeRow
          label={key}
          value={typeof val === "object"
            ? JSON.stringify(val)
            : String(val ?? "")}
        />
      {/each}
    {:else}
      <span class="text-xs text-fg-muted">None</span>
    {/if}
  </DescribeSection>

  <!-- Input Variables -->
  <DescribeSection title="Input">
    {#if inputs.length > 0}
      {#each inputs as v}
        <DescribeRow
          label={v.name ?? ""}
          value={String(v.defaultValue ?? "")}
        />
      {/each}
    {:else}
      <span class="text-xs text-fg-muted">None</span>
    {/if}
  </DescribeSection>

  <!-- Output Variable -->
  {#if spec?.output}
    <DescribeSection title="Output">
      <DescribeRow
        label={spec.output.name ?? ""}
        value={spec.output.type ?? ""}
      />
    </DescribeSection>
  {/if}
</div>
