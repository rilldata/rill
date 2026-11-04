<script lang="ts">
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import ColorRangeSelector from "@rilldata/web-common/features/canvas/inspector/chart/field-config/ColorRangeSelector.svelte";
  import SingleFieldInput from "@rilldata/web-common/features/canvas/inspector/fields/SingleFieldInput.svelte";
  import type { ColorRangeMapping } from "@rilldata/web-common/features/components/charts/types";
  import { DEFAULT_MAP_COLOR_RANGE, type MapColorConfig } from ".";

  type Props = {
    canvasName: string;
    id?: string;
    metricsView: string;
    colorConfig: MapColorConfig | undefined;
    onChange: (updatedConfig: MapColorConfig) => void;
  };

  let {
    canvasName,
    id = "map-color-measure",
    metricsView,
    colorConfig,
    onChange,
  }: Props = $props();

  function handleMeasureSelect(measure: string) {
    onChange({
      colorRange: colorConfig?.colorRange ?? DEFAULT_MAP_COLOR_RANGE,
      ...colorConfig,
      measure,
    });
  }

  function handleColorRangeChange(_property: string, value: ColorRangeMapping) {
    onChange({ measure: "", ...colorConfig, colorRange: value });
  }
</script>

<div class="space-y-2">
  <InputLabel small label="Color" {id} />

  <SingleFieldInput
    {canvasName}
    metricName={metricsView}
    {id}
    type="measure"
    selectedItem={colorConfig?.measure}
    onSelect={handleMeasureSelect}
  />

  <ColorRangeSelector
    {canvasName}
    colorRange={colorConfig?.colorRange}
    onChange={handleColorRangeChange}
    colorRangeConfig={{ enable: true }}
  />
</div>
