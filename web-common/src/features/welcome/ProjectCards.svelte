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
    createRuntimeServiceUnpackEmptyMutation,
    createRuntimeServiceUnpackExampleMutation,
  } from "../../runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { EMPTY_PROJECT_TITLE } from "./constants";
  import { EXAMPLES } from "./constants";
  import { connectorIconMapping } from "@rilldata/web-common/features/connectors/connector-icon-mapping.ts";
  import ProjectCard from "@rilldata/web-common/features/welcome/ProjectCard.svelte";

  const runtimeClient = useRuntimeClient();

  const unpackExampleProject =
    createRuntimeServiceUnpackExampleMutation(runtimeClient);
  const unpackEmptyProject =
    createRuntimeServiceUnpackEmptyMutation(runtimeClient);

  let selectedProjectName: string | null = null;

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
        [key]: selectedProjectName,
        force: true,
      });

      setTimeout(() => {
        window.location.assign("/?redirect=true");
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
      {@const loading = selectedProjectName === example.name}
      <ProjectCard
        onClick={() => unpackProject(example)}
        {loading}
        disabled={!!selectedProjectName}
        label={example.title}
      >
        <svelte:fragment slot="icon">
          {#if icon}
            <svelte:component this={icon} />
          {/if}
        </svelte:fragment>
        <span>{example.title}</span>
      </ProjectCard>
    {/each}

    <ProjectCard
      onClick={() => unpackProject()}
      loading={selectedProjectName === EMPTY_PROJECT_TITLE}
      disabled={!!selectedProjectName}
      label="Start with an empty project"
    >
      <svelte:fragment slot="icon">
        <AddCircleOutline size="16px" />
      </svelte:fragment>
      <span>Start with an empty project</span>
    </ProjectCard>
  </div>
</section>
