<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import ReportIcon from "@rilldata/web-common/components/icons/ReportIcon.svelte";
  import ResourceListRow from "@rilldata/web-common/features/resources/ResourceListRow.svelte";
  import cronstrue from "cronstrue";
  import { PlayIcon, Pencil, Trash2Icon } from "lucide-svelte";
  import ProjectAccessControls from "../../projects/ProjectAccessControls.svelte";
  import { formatRunDate } from "../tableUtils";
  import ReportOwnerBullet from "./ReportOwnerBullet.svelte";

  export let organization: string;
  export let project: string;
  export let id: string;
  export let title: string;
  export let lastRun: string | undefined;
  export let timeZone: string;
  export let frequency: string;
  export let ownerId: string;
  export let lastRunErrorMessage: string | undefined;

  const humanReadableFrequency = cronstrue.toString(frequency);

  let isDropdownOpen = false;
</script>

<ResourceListRow
  href={`reports/${id}`}
  {title}
  icon={ReportIcon}
  errorMessage={lastRunErrorMessage}
>
  <svelte:fragment slot="subtitle">
    {#if !lastRun}
      <span class="shrink-0">Hasn't run yet</span>
    {:else}
      <span class="shrink-0">Last run {formatRunDate(lastRun, timeZone)}</span>
    {/if}
    <span class="shrink-0">•</span>
    <span class="shrink-0 truncate">{humanReadableFrequency}</span>
    <ProjectAccessControls {organization} {project}>
      <svelte:fragment slot="manage-project">
        <span class="shrink-0">•</span>
        <ReportOwnerBullet {organization} {project} {ownerId} />
      </svelte:fragment>
    </ProjectAccessControls>
  </svelte:fragment>

  <svelte:fragment slot="actions">
    <DropdownMenu.Root bind:open={isDropdownOpen}>
      <DropdownMenu.Trigger class="flex-none">
        <IconButton rounded active={isDropdownOpen}>
          <ThreeDot size="16px" />
        </IconButton>
      </DropdownMenu.Trigger>
      <DropdownMenu.Content align="end" class="min-w-[95px]">
        <DropdownMenu.Item class="font-normal flex items-center">
          <PlayIcon size="12px" />
          <span class="ml-2">Run</span>
        </DropdownMenu.Item>
        <DropdownMenu.Item class="font-normal flex items-center">
          <Pencil size="12px" />
          <span class="ml-2">Edit</span>
        </DropdownMenu.Item>
        <DropdownMenu.Item
          class="font-normal flex items-center"
          type="destructive"
        >
          <Trash2Icon size="12px" />
          <span class="ml-2">Delete</span>
        </DropdownMenu.Item>
      </DropdownMenu.Content>
    </DropdownMenu.Root>
  </svelte:fragment>
</ResourceListRow>
