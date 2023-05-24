<script lang="ts">
  import { goto } from "$app/navigation";
  import AddCircleOutline from "@rilldata/web-common/components/icons/AddCircleOutline.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import Card from "../../components/card/Card.svelte";
  import CardHeader from "../../components/card/CardHeader.svelte";
  import { createRuntimeServiceUnpackEmpty } from "../../runtime-client";
  import { EMPTY_PROJECT_TITLE } from "./constants";

  const unpackEmptyProject = createRuntimeServiceUnpackEmpty();
  async function startWithEmptyProject() {
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

<Card on:click={startWithEmptyProject}>
  <AddCircleOutline size="2em" className="text-slate-800" />
  <CardHeader position="middle">Start with an empty project</CardHeader>
</Card>
