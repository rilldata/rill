<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
  } from "@rilldata/web-common/components/dropdown-menu";
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

<DropdownMenu>
  <DropdownMenuTrigger>
    <Button type="secondary">Share</Button>
  </DropdownMenuTrigger>
  <DropdownMenuContent align="end">
    <DropdownMenuItem
      on:click={() => {
        handleCopyLink();
      }}
    >
      <Link size="16px" className="text-gray-900 mr-2 h-4 w-4" />
      Copy shareable link
    </DropdownMenuItem>
  </DropdownMenuContent>
</DropdownMenu>
