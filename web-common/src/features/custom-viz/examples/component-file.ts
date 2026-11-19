import { CUSTOM_VIZ_FOLDER } from "@rilldata/web-common/features/custom-viz/params";
import { getName } from "@rilldata/web-common/features/entity-management/name-utils";

/**
 * Resolves the component name and file path a gallery example imports to,
 * deduped against the components that already exist. Shared by the deterministic
 * and AI import paths so the two can never disagree on naming.
 */
export function componentFilePathFor(
  exampleName: string,
  existingComponentNames: string[],
): { componentName: string; filePath: string } {
  const componentName = getName(
    sanitizeComponentName(exampleName),
    existingComponentNames,
  );
  return {
    componentName,
    filePath: `/${CUSTOM_VIZ_FOLDER}/${componentName}.yaml`,
  };
}

function sanitizeComponentName(name: string): string {
  const sanitized = name
    .toLowerCase()
    .replace(/[^a-z0-9_-]/g, "_")
    .replace(/^(\d)/, "_$1");
  return sanitized || "custom_viz";
}
