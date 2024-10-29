<script lang="ts">
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import { parseRillTimeSyntax } from "@rilldata/web-common/features/dashboards/url-state/time-ranges/rill-time";

  let time = "";

  function parse() {
    parseRillTimeSyntax(time);
  }
</script>

<Input bind:value={time} />
<button on:click={parse}>Parse</button>
