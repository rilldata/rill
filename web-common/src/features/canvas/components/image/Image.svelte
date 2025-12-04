<script lang="ts">
  import ComponentError from "@rilldata/web-common/features/components/ComponentError.svelte";
  import httpClient from "@rilldata/web-common/runtime-client/http-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import ComponentHeader from "../../ComponentHeader.svelte";
  import type { ImageComponent } from "./";
  import { getImagePosition } from "./util";

  export let component: ImageComponent;

  $: ({ specStore } = component);

  $: ({ instanceId } = $runtime);
  $: imageProperties = $specStore;

  $: ({ title, description, show_description_as_tooltip, alignment, url } =
    imageProperties);

  $: objectPosition = getImagePosition(alignment);

  let imageSrc: string | null = null;
  let errorMessage: string | null = null;
  $: {
    if (url) {
      fetchImage(url);
    } else {
      imageSrc = null;
      errorMessage = "No image URL provided";
    }
  }

  async function fetchImage(url: string) {
    try {
      imageSrc = await getImageURL(url);
      errorMessage = null;
    } catch (error) {
      imageSrc = null;
      errorMessage = error.message || "Failed to load image";
    }
  }

  const getImageURL = async (url: string): Promise<string> => {
    if (isValidURL(url)) return url;

    try {
      const response = (await httpClient({
        url: `/v1/instances/${instanceId}/assets/${url}`,
        method: "GET",
      })) as Response;

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
