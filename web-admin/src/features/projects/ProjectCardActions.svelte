<script lang="ts">
  import * as Dropdown from "@rilldata/web-common/components/dropdown-menu";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import FeatherEditIcon from "@rilldata/web-common/components/icons/FeatherEditIcon.svelte";
  import Trash from "@rilldata/web-common/components/icons/Trash.svelte";
  import { ShareIcon } from "lucide-svelte";

  let {
    organization,
    project,
    open = $bindable(false),
    onRename,
    onDelete,
  }: {
    organization: string;
    project: string;
    open?: boolean;
    onRename: () => void;
    onDelete: () => void;
  } = $props();
</script>

<div class="flex flex-row gap-2">
  <Dropdown.Root bind:open>
    <Dropdown.Trigger>
      <ThreeDot size="16px" />
    </Dropdown.Trigger>
    <Dropdown.Content class="w-48" align="start" side="right">
      <Dropdown.Item class="text-sm" onclick={onRename}>
        <FeatherEditIcon /> Rename
      </Dropdown.Item>
      <Dropdown.Item
        href="/{organization}/{project}/-/dashboards?share=true"
        class="text-sm"
      >
        <ShareIcon size={14} /> Share
      </Dropdown.Item>
      <Dropdown.Item class="text-sm text-destructive" onclick={onDelete}>
        <Trash /> Delete
      </Dropdown.Item>
    </Dropdown.Content>
  </Dropdown.Root>
</div>
