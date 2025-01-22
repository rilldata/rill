<script lang="ts">
  import PickOlapConnector from "@rilldata/web-common/features/welcome/wizard/PickOLAPConnector.svelte";
  import PickOlapManagement from "@rilldata/web-common/features/welcome/wizard/PickOLAPManagement.svelte";
  import PickFirstSource from "@rilldata/web-common/features/welcome/wizard/PickSourceConnector.svelte";
  import type { PageData } from "../$types";

  export let data: PageData;

  const { onboardingState } = data;
  const { managementType, olapDriver, firstDataSource } = onboardingState;
</script>

<h1>Let's set up your project</h1>

<div class="wrapper">
  <h2>
    Choose an OLAP database for modeling data and serving dashboards.
    <a
      href="https://docs.rilldata.com/concepts/OLAP"
      target="_blank"
      rel="noopener noreferrer">Learn more</a
    >
  </h2>

  <div>
    <PickOlapManagement {onboardingState} />
    <PickOlapConnector
      {onboardingState}
      continueHref={`/welcome/add-credentials`}
    />
  </div>
</div>

{#if $managementType === "rill-managed"}
  <PickFirstSource
    {onboardingState}
    continueHref={`/welcome/add-credentials`}
    skipHref={"/files/rill.yaml"}
  />
{/if}

<style lang="postcss">
  h1 {
    @apply text-lead text-slate-800;
  }

  h2 {
    @apply text-subheading;
  }

  .wrapper {
    @apply flex flex-col gap-y-4;
  }
</style>
