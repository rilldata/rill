<script lang="ts">
  import * as Dropdown from "@rilldata/web-common/components/dropdown-menu";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import * as m from "@rilldata/web-common/paraglide/messages.js";
  import FeatherEditIcon from "@rilldata/web-common/components/icons/FeatherEditIcon.svelte";
  import PencilIcon from "@rilldata/web-common/components/icons/PencilIcon.svelte";
  import Trash from "@rilldata/web-common/components/icons/Trash.svelte";
  import { ShareIcon } from "lucide-svelte";

  let {
    organization,
    project,
    open = $bindable(false),
    onEdit,
    onRename,
    onDelete,
  }: {
    organization: string;
    project: string;
    open?: boolean;
    onEdit: () => void;
    onRename: () => void;
    onDelete: () => void;
  } = $props();
</script>

<Dropdown.Root bind:open>
  <Dropdown.Trigger>
    <ThreeDot size="16px" />
  </Dropdown.Trigger>
  <Dropdown.Content class="w-48" align="start" side="right">
    <Dropdown.Item class="text-sm" onclick={onEdit}>
      <FeatherEditIcon /> {m.project_edit()}
    </Dropdown.Item>
    <Dropdown.Item class="text-sm" onclick={onRename}>
      <PencilIcon /> {m.project_rename()}
    </Dropdown.Item>
    <Dropdown.Item
      href="/{organization}/{project}/-/dashboards?share=true"
      class="text-sm"
    >
      <ShareIcon size={14} /> {m.project_share()}
    </Dropdown.Item>
    <Dropdown.Item class="text-sm text-destructive" onclick={onDelete}>
      <Trash /> {m.project_delete()}
    </Dropdown.Item>
  </Dropdown.Content>
</Dropdown.Root>
