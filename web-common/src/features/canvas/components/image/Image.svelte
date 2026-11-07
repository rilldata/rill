<script lang="ts">
  import ComponentError from "@rilldata/web-common/features/components/ComponentError.svelte";
  import { themeControl } from "@rilldata/web-common/features/themes/theme-control";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import ComponentHeader from "../../ComponentHeader.svelte";
  import type { ImageComponent } from "./";
  import { getImagePosition } from "./util";

  const client = useRuntimeClient();

  export let component: ImageComponent;

  $: ({ specStore } = component);

  $: ({ instanceId } = client);
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
      const src = await getImageURL(url);
      if (id !== fetchId) return;
      imageSrc = src;
      errorMessage = null;
    } catch (error) {
      if (id !== fetchId) return;
      imageSrc = null;
      errorMessage = error.message || "Failed to load image";
    }
  }

  const getImageURL = async (url: string): Promise<string> => {
    if (isValidURL(url)) return url;

    try {
      const fetchUrl = `${client.host}/v1/instances/${instanceId}/assets/${url}`;
      const headers: Record<string, string> = {};
      const jwt = client.getJwt();
      if (jwt) headers["Authorization"] = `Bearer ${jwt}`;
      const response = await fetch(fetchUrl, { method: "GET", headers });
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      const blob = await response.blob();
      return URL.createObjectURL(blob);
    } catch {
      throw new Error("Failed to fetch image from server");
    }
  };

  const isValidURL = (string: string) => {
    const regex = /^(https?):\/\/[^\s/$.?#].[^\s]*$/i;
    return regex.test(string);
  };
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
