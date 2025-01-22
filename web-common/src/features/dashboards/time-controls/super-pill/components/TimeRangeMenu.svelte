<script lang="ts">
  import { humaniseISODuration } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import type {
    RangeBuckets,
    NamedRange,
    ISODurationString,
  } from "../../new-time-controls";
  import type { CustomEventHandler } from "bits-ui";
  import { RILL_TO_LABEL, ALL_TIME_RANGE_ALIAS } from "../../new-time-controls";
  import SyntaxElement from "./SyntaxElement.svelte";

  export let ranges: RangeBuckets;
  export let selected: NamedRange | ISODurationString;
  export let showDefaultItem: boolean;
  export let defaultTimeRange: NamedRange | ISODurationString | undefined;
  export let onSelectRange: (range: NamedRange | ISODurationString) => void;
  export let onSelectCustomOption: () => void;

  function handleClick(e: CustomEventHandler<MouseEvent, HTMLDivElement>) {
    const range = e.detail.currentTarget.dataset.range;
    if (!range) {
      throw new Error("No range provided");
    }

    onSelectRange(range);
  }
</script>
