import { runtimeServicePutFile } from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { componentFilePathFor } from "./component-file";
import { exampleComponentYAML, type GalleryEntry } from "./example-spec";

/**
 * The I/O side of the gallery import. The conversion itself is pure and lives in
 * example-spec.ts so it can run without the runtime client; the types and loaders
 * are re-exported here for convenience.
 */

export {
  exampleComponentYAML,
  loadGallery,
  type ChannelBinding,
  type GalleryEntry,
  type GalleryIndex,
  type ParamRole,
} from "./example-spec";

/**
 * Imports the chart type as viz_library/<name>.yaml (deduped against existing names)
 * and returns the new file path.
 */
export async function importGalleryChart(
  client: RuntimeClient,
  entry: GalleryEntry,
  existingComponentNames: string[],
): Promise<string> {
  const { filePath } = componentFilePathFor(
    entry.chartSlug,
    existingComponentNames,
  );

  await runtimeServicePutFile(client, {
    path: filePath,
    blob: exampleComponentYAML(entry),
    create: true,
    createOnly: true,
  });

  return filePath;
}
