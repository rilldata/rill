<script lang="ts">
  import { getAbbreviationForIANA } from "@rilldata/web-common/lib/time/timezone";
  import { IANAZone } from "luxon";

  export let watermark = new Date();
  export let iana: string;

  $: zone = new IANAZone(iana);

  $: watermarkTs = watermark.getTime();
</script>

<div class="flex items-center gap-x-1 text-xs cursor-pointer">
  <b class="min-w-9 max-w-9">
    {getAbbreviationForIANA(watermark, iana)}
  </b>

  <p class="inline-block italic min-w-20 max-w-20">
    GMT {zone.formatOffset(watermarkTs, "short")}
  </p>

  <p class="truncate">
    {iana}
  </p>
</div>
