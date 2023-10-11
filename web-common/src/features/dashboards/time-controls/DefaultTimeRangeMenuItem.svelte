<script lang="ts">
  import { humaniseISODuration } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
  import MenuItem from "../../../components/menu/core/MenuItem.svelte";

  export let selected = false;
  export let isoDuration: string;
</script>

<MenuItem animateSelect={false} on:select>
  <div class="w-full flex flex-row flex-grow items-center">
    <div class:font-bold={selected}>
      Last {humaniseISODuration(isoDuration)}
    </div>
  </div>
</MenuItem>
