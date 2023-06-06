<script lang="ts">
  import { goto } from "$app/navigation";
  import AddCircleOutline from "@rilldata/web-common/components/icons/AddCircleOutline.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import Card from "../../components/card/Card.svelte";
  import CardTitle from "../../components/card/CardTitle.svelte";
  import { createRuntimeServiceUnpackEmpty } from "../../runtime-client";
  import { EMPTY_PROJECT_TITLE } from "./constants";
  import { behaviourEvent } from "../../metrics/initMetrics";
  import {
    BehaviourEventAction,
    BehaviourEventMedium,
  } from "../../metrics/service/BehaviourEventTypes";
  import { MetricsEventSpace } from "../../metrics/service/MetricsTypes";

  const unpackEmptyProject = createRuntimeServiceUnpackEmpty();
  async function startWithEmptyProject() {
    behaviourEvent.fireSplashEvent(
      BehaviourEventAction.ProjectEmpty,
      BehaviourEventMedium.Card,
      MetricsEventSpace.Workspace
    );

    $unpackEmptyProject.mutate(
      {
        instanceId: $runtime.instanceId,
        data: {
          title: EMPTY_PROJECT_TITLE,
          force: true,
        },
      },
      {
        onSuccess: () => {
          goto("/");
        },
      }
    );
  }
</script>

<Card
  bgClasses="bg-gradient-to-b from-white to-[#F8FAFC]"
  on:click={startWithEmptyProject}
>
  <AddCircleOutline size="2em" className="text-slate-600" />
  <CardTitle position="middle">Start with an empty project</CardTitle>
</Card>
