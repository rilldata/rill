<script lang="ts">
  import AddCircleOutline from "@rilldata/web-common/components/icons/AddCircleOutline.svelte";
  import Subheading from "@rilldata/web-common/components/typography/Subheading.svelte";
  import { behaviourEvent } from "../../metrics/initMetrics";
  import {
    BehaviourEventAction,
    BehaviourEventMedium,
  } from "../../metrics/service/BehaviourEventTypes";
  import { MetricsEventSpace } from "../../metrics/service/MetricsTypes";
  import {
    createRuntimeServiceUnpackEmpty,
    createRuntimeServiceUnpackExample,
  } from "../../runtime-client";
  import { runtime } from "../../runtime-client/runtime-store";
  import { EMPTY_PROJECT_TITLE } from "./constants";
  import { EXAMPLES } from "./constants";
  import { connectorIconMapping } from "@rilldata/web-common/features/connectors/connector-icon-mapping.ts";
  import { Button } from "@rilldata/web-common/components/button";

  const unpackExampleProject = createRuntimeServiceUnpackExample();
  const unpackEmptyProject = createRuntimeServiceUnpackEmpty();

  let selectedProjectName: string | null = null;

  $: ({ instanceId } = $runtime);

  $: ({ mutateAsync: unpackExample } = $unpackExampleProject);
  $: ({ mutateAsync: unpackEmpty } = $unpackEmptyProject);

  async function unpackProject(example?: (typeof EXAMPLES)[number]) {
    selectedProjectName = example ? example.name : EMPTY_PROJECT_TITLE;

    await behaviourEvent?.fireSplashEvent(
      example
        ? BehaviourEventAction.ExampleAdd
        : BehaviourEventAction.ProjectEmpty,
      BehaviourEventMedium.Card,
      MetricsEventSpace.Workspace,
      example?.name,
    );

    const mutationFunction = example ? unpackExample : unpackEmpty;
    const key = example ? "name" : "displayName";

    try {
      await mutationFunction({
        instanceId,
        data: {
          [key]: selectedProjectName,
          force: true,
        },
      });

      setTimeout(() => {
        if (window.location.search.includes("redirect=true"))
          window.location.reload();
      }, 5000);
    } catch {
      selectedProjectName = null;
    }
  }
</script>

<section class="flex flex-col items-center gap-y-5">
  <Subheading>Or jump right into an example project.</Subheading>
  <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
    {#each EXAMPLES as example (example.name)}
      {@const icon = connectorIconMapping[example.connector]}
      <button on:click={() => unpackProject(example)}>
        {#if icon}
          <svelte:component this={icon} />
        {/if}
        <span>{example.title}</span>
      </button>
    {/each}

    <button on:click={() => unpackProject()}>
      <AddCircleOutline size="2em" />
      <span>Start with an empty project</span>
    </button>
  </div>
</section>

<style lang="postcss">
  button {
    @apply flex flex-row items-center justify-center gap-2 px-4 py-2;
    @apply text-sm bg-surface-overlay rounded-md border text-fg-secondary;
  }
  button:hover {
    @apply border-accent-primary-action shadow-lg cursor-pointer;
  }
</style>
