import { runtimeServicePutFile } from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { componentFilePathFor } from "./component-file";
import { exampleComponentYAML, type GalleryExample } from "./example-spec";

/**
 * The I/O side of the gallery import. The conversion itself is pure and lives in
 * example-spec.ts, so the snapshot script can reuse it without pulling in the
 * runtime client; the types and loaders are re-exported here for convenience.
 */

export {
  exampleComponentYAML,
  loadExample,
  loadGallery,
  parameterizeExampleSpec,
  type GalleryEntry,
  type GalleryExample,
  type GalleryIndex,
} from "./example-spec";

/**
 * Imports the example as viz_library/<name>.yaml (deduped against existing names)
 * and returns the new file path.
 */
export async function importVegaExample(
  client: RuntimeClient,
  example: GalleryExample,
  existingComponentNames: string[],
): Promise<string> {
  const { filePath } = componentFilePathFor(
    example.name,
    existingComponentNames,
  );

  await runtimeServicePutFile(client, {
    path: filePath,
    blob: exampleComponentYAML(example),
    create: true,
    createOnly: true,
  });

  return filePath;
}
