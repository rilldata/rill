<script lang="ts">
  import { BellIcon } from "lucide-svelte";
  import CancelCircleInverse from "@rilldata/web-common/components/icons/CancelCircleInverse.svelte";
  import CheckCircleOutline from "@rilldata/web-common/components/icons/CheckCircleOutline.svelte";
  import ProjectAccessControls from "../../projects/ProjectAccessControls.svelte";
  import AlertOwnerBullet from "./AlertOwnerBullet.svelte";

  export let organization: string;
  export let project: string;
  export let id: string;
  export let title: string;
  export let lastTrigger: string | undefined;
  export let ownerId: string;
  export let lastTriggerErrorMessage: string | undefined;
</script>

<a href={`alerts/${id}`} class="flex flex-col gap-y-0.5 group px-4 py-[5px]">
  <div class="flex gap-x-2 items-center text-slate-500">
    <BellIcon size="14px" />
    <div
      class="text-gray-700 text-sm font-semibold group-hover:text-primary-600"
    >
      {title}
    </div>
    {#if lastTrigger}
      {#if lastTriggerErrorMessage}
        <CancelCircleInverse className="text-red-500" />
      {:else}
        <CheckCircleOutline className="text-primary-500" />
      {/if}
    {/if}
  </div>
  <div class="flex gap-x-1 text-gray-500 text-xs font-normal">
    {#if !lastTrigger}
      <span>Hasn't triggered yet</span>
    {:else}
      <!-- TODO: format -->
      <span>Last triggered {lastTrigger}</span>
    {/if}
    <ProjectAccessControls {organization} {project}>
      <svelte:fragment slot="manage-project">
        <span>•</span>
        <AlertOwnerBullet {organization} {project} {ownerId} />
      </svelte:fragment>
    </ProjectAccessControls>
  </div>
</a>
