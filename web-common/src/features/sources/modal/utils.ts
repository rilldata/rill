import { ConnectorDriverPropertyType } from "@rilldata/web-common/runtime-client";

export function getSpecDefaults(properties: any) {
  const defaults = {};
  (properties ?? []).forEach((property) => {
    if (property.default !== undefined) {
      let value = property.default;
      // Convert to correct type
      if (property.type === ConnectorDriverPropertyType.TYPE_BOOLEAN) {
        value = value === "true";
      }
      defaults[property.key] = value;
    }
  });
  return defaults;
}
