const fs = require('fs');
const yaml = require('js-yaml');
const path = require('path');

// Convert to OpenAPI v3 structure
const v3Spec = {
  openapi: '3.0.3',
  info: {
    title: 'Rill Admin API',
    description: 'Public API endpoints for Rill Admin Service',
    version: '1.0.0'
  },
  servers: [
    {
      url: 'https://admin.rilldata.com',
      description: 'Rill Cloud API'
    }
  ],
  paths: {},
  components: {
    schemas: {},
    securitySchemes: {
      BearerAuth: {
        type: 'http',
        scheme: 'bearer'
      }
    }
  },
  security: [
    {
      BearerAuth: []
    }
  ]
};

function extractPublicMethods(protoDir) {
  const publicMethods = new Set();
  const apiProtoPath = path.join(protoDir, 'rill/admin/v1/api.proto');

  if (fs.existsSync(apiProtoPath)) {
    const content = fs.readFileSync(apiProtoPath, 'utf8');

    const rpcRegex = /rpc\s+(\w+)\s*\([^)]+\)\s*returns\s*\([^)]+\)\s*\{([^}]+)\}/g;
    let match;

    while ((match = rpcRegex.exec(content)) !== null) {
      const [, methodName, options] = match;
      if (options.includes('method_visibility') && options.includes('PUBLIC')) {
        publicMethods.add(methodName);
      }
    }
  }

  return publicMethods;
}

// Convert OpenAPI v2 to v3 and filter public methods
function convertAndFilterSpec(inputFile, outputFile, publicMethods) {
  const spec = yaml.load(fs.readFileSync(inputFile, 'utf8'));

  for (const [path, pathItem] of Object.entries(spec.paths || {})) {
    const filteredPathItem = {};

    for (const [method, operation] of Object.entries(pathItem)) {
      if (typeof operation !== 'object' || !operation.operationId) continue;

      const methodName = operation.operationId.replace('AdminService_', '');

      if (publicMethods.has(methodName)) {
        const v3Operation = {
          ...operation,
          requestBody: convertRequestBody(operation),
          responses: convertResponses(operation.responses || {})
        };

        if (v3Operation.parameters) {
          v3Operation.parameters = v3Operation.parameters
            .filter(p => p.in !== 'body')
            .map(p => ({
              ...p,
              schema: p.schema ? convertSchema(p.schema) : p.schema
            }));
        }

        delete v3Operation.consumes;
        delete v3Operation.produces;

        filteredPathItem[method] = v3Operation;
      }
    }

    if (Object.keys(filteredPathItem).length > 0) {
      v3Spec.paths[path] = filteredPathItem;
    }
  }

  const usedSchemas = new Set();
  collectUsedSchemas(v3Spec.paths, usedSchemas, spec.definitions);

  for (const schemaName of usedSchemas) {
    if (spec.definitions?.[schemaName]) {
      v3Spec.components.schemas[schemaName] = convertSchema(spec.definitions[schemaName]);
    }
  }

  fs.writeFileSync(outputFile, yaml.dump(v3Spec, { lineWidth: -1 }));
}

function convertRequestBody(operation) {
  const bodyParam = operation.parameters?.find(p => p.in === 'body');
  if (!bodyParam) return undefined;

  return {
    required: true,
    content: {
      'application/json': {
        schema: convertSchema(bodyParam.schema)
      }
    }
  };
}

function convertResponses(v2Responses) {
  const v3Responses = {};

  for (const [code, response] of Object.entries(v2Responses)) {
    v3Responses[code] = {
      description: response.description || '',
      content: response.schema ? {
        'application/json': {
          schema: convertSchema(response.schema)
        }
      } : undefined
    };
  }

  return v3Responses;
}

function convertSchema(schema) {
  if (schema.$ref) {
    return { $ref: schema.$ref.replace('#/definitions/', '#/components/schemas/') };
  }

  if (typeof schema !== 'object' || schema === null) {
    return schema;
  }

  const converted = { ...schema };

  for (const [key, value] of Object.entries(converted)) {
    if (value && typeof value === 'object') {
      if (Array.isArray(value)) {
        converted[key] = value.map(item => convertSchema(item));
      } else {
        converted[key] = convertSchema(value);
      }
    }
  }

  return converted;
}

function collectUsedSchemas(paths, usedSchemas, allDefinitions) {
  const visited = new Set();

  collectSchemasFromObject(paths, usedSchemas);

  let foundNew = true;
  while (foundNew) {
    foundNew = false;
    const currentSchemas = new Set(usedSchemas);

    for (const schemaName of currentSchemas) {
      if (!visited.has(schemaName) && allDefinitions[schemaName]) {
        visited.add(schemaName);
        const prevSize = usedSchemas.size;
        collectSchemasFromObject(allDefinitions[schemaName], usedSchemas);
        if (usedSchemas.size > prevSize) {
          foundNew = true;
        }
      }
    }
  }
}

function collectSchemasFromObject(obj, usedSchemas) {
  const schemaRefRegex = /#\/(?:definitions|components\/schemas)\/([^"'\s]+)/g;
  const objStr = JSON.stringify(obj);
  let match;

  while ((match = schemaRefRegex.exec(objStr)) !== null) {
    usedSchemas.add(match[1]);
  }
}

// Main execution
function main() {
  const protoDir = path.join(__dirname, '../proto');
  const apiDir = path.join(__dirname, '../docs/api/rill/admin/v1');

  const publicMethods = extractPublicMethods(protoDir);

  const inputFile = path.join(apiDir, 'public.swagger.yaml');
  const outputFile = path.join(apiDir, 'public.swagger.yaml');

  if (!fs.existsSync(inputFile)) {
    console.error(`Input file not found: ${inputFile}`);
    process.exit(1);
  }

  convertAndFilterSpec(inputFile, outputFile, publicMethods);
}

if (require.main === module) {
  main();
}

module.exports = { extractPublicMethods, convertAndFilterSpec };
