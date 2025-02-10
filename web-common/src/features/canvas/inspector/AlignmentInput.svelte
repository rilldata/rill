<script lang="ts">
  import IconSwitcher from "@rilldata/web-common/components/forms/IconSwitcher.svelte";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import { defaultAlignment } from "@rilldata/web-common/features/canvas/components/markdown";
  import type {
    ComponentAlignment,
    HoritzontalAlignment,
    VeriticalAlignment,
  } from "@rilldata/web-common/features/canvas/components/types";
  import {
    AlignCenterHorizontal,
    AlignCenterVertical,
    AlignEndHorizontal,
    AlignEndVertical,
    AlignStartHorizontal,
    AlignStartVertical,
  } from "lucide-svelte";

  export let key: string;
  export let label: string;
  export let position: ComponentAlignment | undefined;
  export let onChange: (updatedPosition: ComponentAlignment) => void;

  $: if (position === undefined) {
    position = defaultAlignment;
  }

  const horizontalOptions = [
    { id: "left", Icon: AlignStartVertical },
    { id: "center", Icon: AlignCenterVertical },
    { id: "right", Icon: AlignEndVertical },
  ];

  const verticalOptions = [
    { id: "top", Icon: AlignStartHorizontal },
    { id: "middle", Icon: AlignCenterHorizontal },
    { id: "bottom", Icon: AlignEndHorizontal },
  ];

  const updatePosition = (
    value: string,
    direction: "horizontal" | "vertical",
  ) => {
    if (!position) position = defaultAlignment;
    if (direction === "horizontal") {
      const newOption = value as HoritzontalAlignment;
      if (newOption === position?.horizontal) return;
      onChange({ ...position, horizontal: newOption });
    } else {
      const newOption = value as VeriticalAlignment;
      if (newOption === position?.vertical) return;
      onChange({ ...position, vertical: newOption });
    }
  };
</script>

<div class="flex flex-col gap-y-2">
  <InputLabel small {label} id={key} />

  <IconSwitcher
    small
    expand
    fields={horizontalOptions}
    selected={position?.horizontal}
    onClick={(option) => updatePosition(option, "horizontal")}
  />

  <IconSwitcher
    small
    expand
    fields={verticalOptions}
    selected={position?.vertical}
    onClick={(option) => updatePosition(option, "vertical")}
  />
</div>
