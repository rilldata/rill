<script lang="ts">
  import ComponentError from "@rilldata/web-common/features/components/ComponentError.svelte";
  import { themeControl } from "@rilldata/web-common/features/themes/theme-control";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { onDestroy } from "svelte";
  import ComponentHeader from "../../ComponentHeader.svelte";
  import { ProjectAssetLoader } from "../project-assets";
  import type { ImageComponent } from "./";
  import { getImagePosition } from "./util";

  // A URL that isn't absolute is a project asset, such as `public/img/logo.svg`.
  const assets = new ProjectAssetLoader(useRuntimeClient());
  onDestroy(() => assets.destroy());

  export let component: ImageComponent;

  $: ({ specStore } = component);

  $: imageProperties = $specStore;

  $: ({
    title,
    description,
    show_description_as_tooltip,
    alignment,
    url,
    dark_url,
  } = imageProperties);

  $: objectPosition = getImagePosition(alignment);

  // The image location for the workbook export; see BaseCanvasComponent.exportPayload.
  $: component.exportPayload.set(url ? { kind: "imageUrl", url } : undefined);

  // Prefer the dark mode image when the app is in dark mode; fall back to `url`.
  $: activeUrl = $themeControl === "dark" && dark_url ? dark_url : url;

  let imageSrc: string | null = null;
  let errorMessage: string | null = null;
  // Incremented on every fetch so a slow response for a previous URL
  // (e.g. after toggling the theme) can't overwrite the latest image.
  let fetchId = 0;
  $: {
    if (activeUrl) {
      void fetchImage(activeUrl);
    } else {
      imageSrc = null;
      errorMessage = "No image URL provided";
    }
  }

  async function fetchImage(url: string) {
    const id = ++fetchId;
    try {
      const src = await assets.resolve(url);
      if (id !== fetchId) return;
      imageSrc = src;
      errorMessage = null;
    } catch {
      if (id !== fetchId) return;
      imageSrc = null;
      errorMessage = "Failed to fetch image from server";
    }
  }
</script>

{#if errorMessage}
  <ComponentError error={errorMessage} />
{:else}
  <ComponentHeader
    {component}
    {title}
    {description}
    showDescriptionAsTooltip={show_description_as_tooltip}
  />
  <img
    src={imageSrc || ""}
    alt={"Canvas Image"}
    draggable="false"
    class="h-full w-full overflow-hidden object-contain"
    style:object-position={objectPosition}
  />
{/if}
