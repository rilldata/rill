import type { OpenAPIObject } from "openapi3-ts";

export const INSTANCE_ID_PLACEHOLDER = "INSTANCE_ID";

function transformer(spec: OpenAPIObject): OpenAPIObject {
  const newSpec: OpenAPIObject = { ...spec };

  for (const path in spec.paths) {
    // Instead of using an OpenAPI path parameter for "instanceId", which would require that every calling
    // component provide the variable in its request, we inject a placeholder string, which we can then process
    // later in the http client's request interceptor
    const transformedPath = path.replace(
      "{instanceId}",
      INSTANCE_ID_PLACEHOLDER
    );
    if (transformedPath !== path) {
      newSpec.paths[transformedPath] = newSpec.paths[path];
      delete newSpec.paths[path];
    }
  }

  return newSpec;
}

export default transformer;
