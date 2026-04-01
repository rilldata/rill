<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import { ExternalLinkIcon } from "lucide-svelte";

  export let href: string;
  export let isDropdownOpen: boolean;
  export let onDropdownOpenChange: (isOpen: boolean) => void;
</script>

<DropdownMenu.Root open={isDropdownOpen} onOpenChange={onDropdownOpenChange}>
  <DropdownMenu.Trigger class="flex-none" aria-label="Project actions">
    <IconButton rounded active={isDropdownOpen} size={20}>
      <ThreeDot size="16px" />
    </IconButton>
  </DropdownMenu.Trigger>
  <DropdownMenu.Content align="start">
    <DropdownMenu.Item class="font-normal flex items-center" {href}>
      <div class="flex items-center">
        <ExternalLinkIcon size="12px" />
        <span class="ml-2">View project status</span>
      </div>
    </DropdownMenu.Item>
  </DropdownMenu.Content>
</DropdownMenu.Root>
