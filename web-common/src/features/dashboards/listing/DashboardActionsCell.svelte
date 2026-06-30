<script lang="ts">
  import { goto } from "$app/navigation";
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import { BellPlusIcon, MailPlus, Share2 } from "lucide-svelte";

  export let dashboardHref: string;
  export let title: string;
  export let isMetricsExplorer: boolean;

  let isDropdownOpen = false;

  // The Create Alert / Create Report / Share dialogs live on the dashboard
  // pages and depend on dashboard state managers. From the listing page we
  // navigate to the dashboard with a query param so the page can auto-open
  // the relevant dialog/popover.
  function openCreateAlert() {
    void goto(`${dashboardHref}?action=create-alert`);
  }

  // Reports start from an empty pivot so the user can shape the table data
  // before scheduling. `view=pivot&rows=&cols=` clears any persisted pivot
  // state, then `action=create-report` triggers the dialog on arrival.
  function openCreateReport() {
    void goto(`${dashboardHref}?view=pivot&rows=&cols=&action=create-report`);
  }

  function openShare() {
    void goto(`${dashboardHref}?action=share`);
  }
</script>

<div class="flex justify-end" data-no-row-click>
  <DropdownMenu.Root bind:open={isDropdownOpen}>
    <DropdownMenu.Trigger class="flex-none">
      <IconButton
        rounded
        active={isDropdownOpen}
        ariaLabel={`Actions for ${title}`}
      >
        <ThreeDot size="16px" />
      </IconButton>
    </DropdownMenu.Trigger>
    <DropdownMenu.Content side="bottom" align="start" class="min-w-[160px]">
      <DropdownMenu.Item
        class="font-normal flex items-center"
        onclick={openShare}
      >
        <Share2 size="12px" />
        <span class="ml-2">Share</span>
      </DropdownMenu.Item>
      {#if isMetricsExplorer}
        <DropdownMenu.Item
          class="font-normal flex items-center"
          onclick={openCreateAlert}
        >
          <BellPlusIcon size="12px" />
          <span class="ml-2">Create alert</span>
        </DropdownMenu.Item>
        <DropdownMenu.Item
          class="font-normal flex items-center"
          onclick={openCreateReport}
        >
          <MailPlus size="12px" />
          <span class="ml-2">Create report</span>
        </DropdownMenu.Item>
      {/if}
    </DropdownMenu.Content>
  </DropdownMenu.Root>
</div>
