<script lang="ts">
  import { SidebarCTAButton } from "@rilldata/web-common/components/button";
  import WithTogglableFloatingElement from "@rilldata/web-common/components/floating-element/WithTogglableFloatingElement.svelte";
  import Add from "@rilldata/web-common/components/icons/Add.svelte";
  import { LIST_SLIDE_DURATION } from "@rilldata/web-local/lib/application-config";
  import { fly } from "svelte/transition";
  import CreateNewDashboardMenu from "./CreateNewDashboardMenu.svelte";
</script>

<div class="px-4">
  <WithTogglableFloatingElement
    location="right"
    alignment="start"
    distance={16}
    let:active
    let:toggleFloatingElement={toggle}
  >
    <SidebarCTAButton on:click={toggle}
      >Create Dashboard <Add /></SidebarCTAButton
    >
    <!-- <button
      class:bg-gray-100={active}
      class:bg-gray-50={!active}
      class="w-full rounded flex items-center gap-x-2 justify-center py-1 border hover:bg-gray-100 text-gray-500"
      on:click={toggle}
    >
      create a dashboard <Add />
    </button> -->

    <div
      slot="floating-element"
      transition:fly={{ duration: LIST_SLIDE_DURATION / 2, x: -4 }}
    >
      <CreateNewDashboardMenu
        on:item-select={toggle}
        on:escape={toggle}
        on:click-outside={toggle}
        on:create-empty-dashboard
        on:bootstrap-dashboard
      />
    </div>
  </WithTogglableFloatingElement>
</div>
