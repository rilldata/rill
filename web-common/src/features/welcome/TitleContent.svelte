<script lang="ts">
  import { goto } from "$app/navigation";
  import Add from "@rilldata/web-common/components/icons/Add.svelte";
  import RillLogoSquareNegative from "@rilldata/web-common/components/icons/RillLogoSquareNegative.svelte";
  import RadixH1 from "@rilldata/web-common/components/typography/RadixH1.svelte";
  import Subheading from "@rilldata/web-common/components/typography/Subheading.svelte";
  import AddSourceModal from "@rilldata/web-common/features/sources/add-source/AddSourceModal.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { createRuntimeServiceUnpackEmpty } from "../../runtime-client";
  import { EMPTY_PROJECT_TITLE } from "./constants";

  let showAddSourceModal = false;
  const openShowAddSourceModal = () => {
    showAddSourceModal = true;
  };

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

<section class="flex flex-col gap-y-6 items-center text-center">
  <RillLogoSquareNegative size="84px" gradient />
  <RadixH1>
    <span
      class="bg-gradient-to-r from-[#4D488C] to-[#515F9D] text-transparent bg-clip-text"
    >
      Welcome to Rill
    </span>
  </RadixH1>
  <div class="flex flex-col gap-y-2">
    <Subheading twColor="text-slate-600">
      A Rill project lets you build an <span class="text-base font-semibold"
        >un-dashboard</span
      > that people will actually use.
    </Subheading>
    <Subheading twColor="text-slate-600">Let’s get started.</Subheading>
  </div>
  <button
    class="pl-2 pr-4 py-2 rounded-sm bg-gradient-to-b from-[#4680FF] to-[#2563EB] hover:from-blue-500 hover:to-blue-500"
    on:click={openShowAddSourceModal}
  >
    <div
      class="flex flex-row gap-x-1 items-center text-sm font-medium text-white"
    >
      <Add className="text-white" />
      Add data
    </div>
  </button>
  <button
    class="px-2 font-medium text-xs text-blue-500 hover:text-blue-400"
    on:click={startWithEmptyProject}
  >
    Or start with an empty project
  </button>
  {#if showAddSourceModal}
    <AddSourceModal
      on:close={() => {
        showAddSourceModal = false;
      }}
    />
  {/if}
</section>
