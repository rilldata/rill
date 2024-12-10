<script lang="ts">
  import type { ImageProperties } from "@rilldata/web-common/features/templates/types";
  import type { V1ComponentSpecRendererProperties } from "@rilldata/web-common/runtime-client";
  import httpClient from "@rilldata/web-common/runtime-client/http-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { get } from "svelte/store";

  export let rendererProperties: V1ComponentSpecRendererProperties;

  const instanceId = get(runtime).instanceId;
  const DEFAULT_IMAGE_PROPERTIES: ImageProperties = {
    url: "",
    css: {
      "object-fit": "contain",
      opacity: "1",
      filter: "blur(0px) saturate(1)",
    },
  };

  $: imageProperties = {
    ...DEFAULT_IMAGE_PROPERTIES,
    ...rendererProperties,
    css: {
      ...DEFAULT_IMAGE_PROPERTIES.css,
      ...rendererProperties.css,
    },
  } as ImageProperties;

  $: styleString = Object.entries(imageProperties.css || {})
    .map(([k, v]) => `${k}:${v}`)
    .join(";");

  let imageSrc: string | null = null;
  let errorMessage: string | null = null;
  $: {
    if (imageProperties.url) {
      fetchImage(imageProperties.url);
    } else {
      imageSrc = null;
      errorMessage = null;
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
  <div class="error-message">{errorMessage}</div>
{:else}
  <img
    src={imageSrc || ""}
    alt={"Dashboard Image"}
    draggable="false"
    style={styleString}
  />
{/if}

<style>
  .error-message {
    color: red;
    font-weight: bold;
  }
</style>
