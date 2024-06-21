<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import Link from "@rilldata/web-common/components/icons/Link.svelte";
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/copy-to-clipboard";
</script>

<DropdownMenu.Root>
  <DropdownMenu.Trigger asChild let:builder>
    <Button type="secondary" builders={[builder]}>Share</Button>
  </DropdownMenu.Trigger>
  <DropdownMenu.Content align="end">
    <DropdownMenu.Item
      on:click={() => {
        copyToClipboard(window.location.href, "Link copied to clipboard");
      }}
    >
      <Link size="16px" className="text-gray-900 mr-2 h-4 w-4" />
      Copy shareable link
    </DropdownMenu.Item>
  </DropdownMenu.Content>
</DropdownMenu.Root>
