<script lang="ts">
  import { ImageProperties } from "@rilldata/web-common/features/templates/types";
  import { V1ComponentSpecRendererProperties } from "@rilldata/web-common/runtime-client";

  export let rendererProperties: V1ComponentSpecRendererProperties;

  // Define default values for image properties
  const defaultImageProperties: ImageProperties = {
    url: "",
    size: "contain",
    adjust: {
      opacity: 1,
      blur: 0,
      saturation: 1,
    },
  };

  // Merge defaults with provided rendererProperties
  $: imageProperties = {
    ...defaultImageProperties,
    ...rendererProperties,
    adjust: {
      ...defaultImageProperties.adjust,
      ...rendererProperties.adjust,
    },
  } as ImageProperties;
</script>

<img
  src={imageProperties.url}
  alt="Dashboard Image"
  style={`
    object-fit: ${imageProperties.size}; 
    opacity: ${imageProperties.adjust?.opacity}; 
    filter: blur(${imageProperties.adjust?.blur}px) saturate(${imageProperties.adjust?.saturation});
    `}
/>
