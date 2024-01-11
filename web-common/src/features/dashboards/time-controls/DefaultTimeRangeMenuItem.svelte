<script lang="ts">
  import { humaniseISODuration } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
  import DropdownMenuItem from "@rilldata/web-common/components/dropdown-menu/DropdownMenuItem.svelte";

  export let selected = false;
  export let isoDuration: string;
</script>

<DropdownMenuItem on:click on:on:mouseenter>
  <div class="w-full flex flex-row flex-grow items-center">
    <div class:font-bold={selected}>
      Last {humaniseISODuration(isoDuration)}
    </div>
  </div>
</DropdownMenuItem>
