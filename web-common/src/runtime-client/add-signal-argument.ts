import ts from "typescript";
import orvalConfig from "../../orval.config";
import * as prettier from "prettier";
import { writeFile } from "node:fs/promises";

/*
 * Orval is not generating code for POST requests as expected.
 * We have a few POST queries that are essentially GET requests with large request in the POST body.
 * For orval to generate a query we set useQuery and signal. Signal allows for query cancellation.
 * But orval only honours signal in the function creating the query object but not the function that makes a call to our http client.
 * This file rewrites the function making the call to http client and adds the signal argument.
 *
 * This transformer takes the output from orval and does in-place rewrite.
 * WARNING: There might be undefined behaviour when used on other files.
 *
 * EG: Consider this function for MetricsViewAggregation,
 *
 * export const queryServiceMetricsViewAggregation = (
 *   instanceId: string,
 *   metricsView: string,
 *   queryServiceMetricsViewAggregationBody: QueryServiceMetricsViewAggregationBody,
 * ) => {
 *   return httpClient<V1MetricsViewAggregationResponse>({
 *     url: `/v1/instances/${instanceId}/queries/metrics-views/${metricsView}/aggregation`,
 *     method: "POST",
 *     headers: { "Content-Type": "application/json" },
 *     data: queryServiceMetricsViewAggregationBody,
 *   });
 * };
 *
 * After running this transformer we get,
 *
 * export const queryServiceMetricsViewAggregation = (
 *   instanceId: string,
 *   metricsView: string,
 *   queryServiceMetricsViewAggregationBody: QueryServiceMetricsViewAggregationBody,
 *   signal?: AbortSignal,
 * ) => {
 *   return httpClient<V1MetricsViewAggregationResponse>({
 *     url: `/v1/instances/${instanceId}/queries/metrics-views/${metricsView}/aggregation`,
 *     method: "POST",
 *     headers: { "Content-Type": "application/json" },
 *     data: queryServiceMetricsViewAggregationBody,
 *     signal,
 *   });
 * };
 */

const Operations: Record<
  string,
  {
    query: {
      useQuery: boolean;
      signal: boolean;
    };
  }
> = (orvalConfig as any).api.output.override.operations;

async function transformFile(fileName: string) {
  const program = ts.createProgram([fileName], {
    moduleResolution: ts.ModuleResolutionKind.Node10,
  });

  let sourceFile = program.getSourceFile(fileName);
  if (!sourceFile) return;
  // Typescript parser doesn't retain blank lines.
  // So we replace those with a comment and add back the blank line after rewriting
  sourceFile = replaceBlankLines(sourceFile);

  const transformationResult = ts.transform(sourceFile, [addSignalTransformer]);
  const transformedSourceFile = transformationResult.transformed[0];

  const printer = ts.createPrinter();
  const result = printer.printNode(
    ts.EmitHint.Unspecified,
    transformedSourceFile,
    sourceFile,
  );

  // Run prettier
  const newCode = await prettier.format(result, {
    parser: "typescript",
  });
  await writeFile(fileName, addBackBlankLines(newCode));
}

function addSignalTransformer(context: ts.TransformationContext) {
  return (rootNode: ts.Node) => {
    function visit(node: ts.Node): ts.Node {
      node = ts.visitEachChild(node, visit, context);

      if (
        // ignore non variable declarations
        !ts.isVariableDeclaration(node) ||
        // ignore non identifier names
        !ts.isIdentifier(node.name) ||
        // ignore methods that are not the query function
        !node.name.escapedText.toString().startsWith("queryService")
      ) {
        return node;
      }

      const queryName = node.name.escapedText.toString();
      if (!isOverriddenPostQuery(queryName)) {
        return node;
      }

      const init = node.initializer;
      // safeguard to make sure initializer is defined and is an arrow function
      if (!init || !ts.isArrowFunction(init) || !ts.isBlock(init.body)) {
        return node;
      }

      const callStatement = init.body.statements[0];
      // some other safeguards to make sure the arguments are as expected
      if (
        !callStatement ||
        !ts.isReturnStatement(callStatement) ||
        !callStatement.expression ||
        !ts.isCallExpression(callStatement.expression) ||
        !callStatement.expression.arguments[0] ||
        !ts.isObjectLiteralExpression(callStatement.expression.arguments[0])
      ) {
        return node;
      }

      const callArg = callStatement.expression.arguments[0];
      const lastProp = callArg.properties[callArg.properties.length - 1];
      // make sure to not add if signal is already present
      if (lastProp?.name && (lastProp.name as any).escapedText === "signal") {
        return node;
      }

      return context.factory.createVariableDeclaration(
        queryName,
        node.exclamationToken,
        node.type,
        context.factory.createArrowFunction(
          init.modifiers,
          init.typeParameters,
          [
            ...init.parameters,
            // add the additional signal param
            context.factory.createParameterDeclaration(
              undefined,
              undefined,
              "signal",
              context.factory.createToken(ts.SyntaxKind.QuestionToken),
              context.factory.createTypeReferenceNode("AbortSignal"),
            ),
          ],
          init.type,
          init.equalsGreaterThanToken,
          context.factory.createBlock([
            context.factory.createReturnStatement(
              context.factory.createCallExpression(
                callStatement.expression.expression,
                callStatement.expression.typeArguments,
                [
                  context.factory.createObjectLiteralExpression(
                    [
                      ...callArg.properties,
                      // add the additional signal property to the argument
                      context.factory.createShorthandPropertyAssignment(
                        "signal",
                      ),
                    ],
                    true,
                  ),
                ],
              ),
            ),
          ]),
        ),
      );
    }

    return ts.visitNode(rootNode, visit);
  };
}

function isOverriddenPostQuery(name: string) {
  const operationName =
    "QueryService_" +
    name.replace(/queryService(.)/, (_, c: string) => c.toUpperCase());
  return Operations[operationName]?.query?.signal;
}

function replaceBlankLines(source: ts.SourceFile) {
  const newCode = source.text.replace(
    /^(\s*)$/gm,
    (_, spaces: string) => spaces + "//__Dummy__",
  );
  return source.update(newCode, {
    span: {
      start: 0,
      length: source.text.length,
    },
    newLength: newCode.length,
  });
}

function addBackBlankLines(code: string) {
  return code.replace(/^\s*\/\/__Dummy__$/gm, () => "");
}

transformFile(process.argv[2]).catch(console.error);
