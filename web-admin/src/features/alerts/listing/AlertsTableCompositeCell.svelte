<script lang="ts">
  import AlertIcon from "@rilldata/web-common/components/icons/AlertIcon.svelte";
  import CancelCircleInverse from "@rilldata/web-common/components/icons/CancelCircleInverse.svelte";
  import CheckCircleOutline from "@rilldata/web-common/components/icons/CheckCircleOutline.svelte";
  import { resourceColorMapping } from "@rilldata/web-common/features/entity-management/resource-icon-mapping";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { timeAgo } from "../../dashboards/listing/utils";
  import ProjectAccessControls from "../../projects/ProjectAccessControls.svelte";
  import AlertOwnerBullet from "./AlertOwnerBullet.svelte";

  export let organization: string;
  export let project: string;
  export let id: string;
  export let title: string;
  export let lastTrigger: string | undefined;
  export let ownerId: string;
  export let lastTriggerErrorMessage: string | undefined;

  const alertColor = resourceColorMapping[ResourceKind.Alert];
</script>

<a
  href={`alerts/${id}`}
  class="flex flex-col gap-y-1 group px-4 py-2.5 w-full h-full"
>
  <div class="flex gap-x-2 items-center min-h-[20px]">
    <AlertIcon size="14px" color={alertColor} />
    <span
      class="text-gray-700 text-sm font-semibold group-hover:text-primary-600 truncate"
    >
      {title}
    </span>
    {#if lastTrigger}
      {#if lastTriggerErrorMessage}
        <CancelCircleInverse className="text-red-500 shrink-0" />
      {:else}
        <CheckCircleOutline className="text-primary-500 shrink-0" />
      {/if}
    {/if}
  </div>
  <div
    class="flex gap-x-1 text-gray-500 text-xs font-normal min-h-[16px] overflow-hidden"
  >
    {#if !lastTrigger}
      <span class="shrink-0">Hasn't been checked yet</span>
    {:else}
      <span class="shrink-0">Last checked {timeAgo(new Date(lastTrigger))}</span
      >
    {/if}
    <ProjectAccessControls {organization} {project}>
      <svelte:fragment slot="manage-project">
        <span class="shrink-0">•</span>
        <AlertOwnerBullet {organization} {project} {ownerId} />
      </svelte:fragment>
    </ProjectAccessControls>
  </div>
</a>
