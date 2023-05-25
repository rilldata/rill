<script lang="ts">
  import { goto } from "$app/navigation";
  import AddCircleOutline from "@rilldata/web-common/components/icons/AddCircleOutline.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";
  import Card from "../../components/card/Card.svelte";
  import CardTitle from "../../components/card/CardTitle.svelte";
  import {
    createRuntimeServiceUnpackEmpty,
    getRuntimeServiceGetFileQueryKey,
  } from "../../runtime-client";
  import { EMPTY_PROJECT_TITLE } from "./constants";

  const queryClient = useQueryClient();

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
          // Invalidate `rill.yaml` GetFile
          queryClient.invalidateQueries({
            queryKey: getRuntimeServiceGetFileQueryKey(
              $runtime.instanceId,
              "rill.yaml"
            ),
          });
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
