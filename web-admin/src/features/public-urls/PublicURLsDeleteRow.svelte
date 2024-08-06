<script lang="ts">
  import { adminServiceRevokeMagicAuthToken } from "@rilldata/web-admin/client";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import Trash from "@rilldata/web-common/components/icons/Trash.svelte";

  export let id: string;

  async function handleClick(e: MouseEvent) {
    e.preventDefault();
    e.stopPropagation();

    adminServiceRevokeMagicAuthToken(id);

    eventBus.emit("notification", {
      message: `Magic auth token deleted`,
    });

    // TODO: refetch queries of public urls
  }
</script>

<div class="flex items-center justify-center">
  <button on:click={handleClick} class="text-gray-400 hover:text-gray-500">
    <Trash size="16px " />
  </button>
</div>
