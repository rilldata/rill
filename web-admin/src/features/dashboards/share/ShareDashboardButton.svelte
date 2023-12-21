<script lang="ts">
  import { buttonClasses } from "@rilldata/web-common/components/button/classes";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import Link from "@rilldata/web-common/components/icons/Link.svelte";
  import { notifications } from "@rilldata/web-common/components/notifications";

  function handleCopyLink() {
    // Copy the current URL to the clipboard
    navigator.clipboard.writeText(window.location.href);

    notifications.send({
      message: "Link copied to clipboard",
    });
  }
</script>

<DropdownMenu.Root>
  <DropdownMenu.Trigger class={buttonClasses({ type: "secondary" })}>
    Share
  </DropdownMenu.Trigger>
  <DropdownMenu.Content align="end">
    <DropdownMenu.Item
      on:click={() => {
        handleCopyLink();
      }}
    >
      <Link size="16px" className="text-gray-900 mr-2 h-4 w-4" />
      Copy shareable link
    </DropdownMenu.Item>
  </DropdownMenu.Content>
</DropdownMenu.Root>
