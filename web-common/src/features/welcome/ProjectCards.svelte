<script lang="ts">
  import AddCircleOutline from "@rilldata/web-common/components/icons/AddCircleOutline.svelte";
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
  import { connectorIconMapping } from "@rilldata/web-common/features/connectors/connector-metadata.ts";
  import ProjectCard from "@rilldata/web-common/features/welcome/ProjectCard.svelte";
  import { splitFolderFileNameAndExtension } from "@rilldata/web-common/features/entity-management/file-path-utils.ts";

  export let skipNavigation = false;
  export let allowEmpty = true;
  export let onSelect: (firstDashboard?: string) => void = () => {};

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

      if (skipNavigation) {
        if (example?.firstFile) {
          const [, dashboardName] = splitFolderFileNameAndExtension(
            example.firstFile,
          );
          onSelect(dashboardName);
        } else {
          onSelect();
        }
        return;
      }

      // TODO: improve navigation in rill dev to avoid this artificial 5 second delay
      setTimeout(() => {
        window.location.assign("/?redirect=true");
      }, 5000);
    } catch {
      selectedProjectName = null;
    }
  }
</script>

<section class="flex flex-col items-center">
  <div class="flex md:flex-row flex-col gap-4">
    {#each EXAMPLES as example (example.name)}
      {@const icon = connectorIconMapping[example.connector]}
      {@const loading = selectedProjectName === example.name}
      <ProjectCard
        onclick={async () => {
          await unpackProject(example);
        }}
        {loading}
        disabled={!!selectedProjectName}
        label={example.title}
      >
        <svelte:fragment slot="icon">
          {#if icon}
            <svelte:component this={icon} size="16px" />
          {/if}
        </svelte:fragment>
        <span>{example.title}</span>
      </ProjectCard>
    {/each}

    {#if allowEmpty}
      <ProjectCard
        onclick={() => unpackProject()}
        loading={selectedProjectName === EMPTY_PROJECT_TITLE}
        disabled={!!selectedProjectName}
        label="Start a blank project"
      >
        <svelte:fragment slot="icon">
          <AddCircleOutline size="16px" />
        </svelte:fragment>
        <span>Start a blank project</span>
      </ProjectCard>
    {/if}
  </div>
</section>
