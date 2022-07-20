import * as ts from "typescript";
import { sync as globSync } from "glob";
import path from "node:path";
import fs from "node:fs";

function getConfig(): ts.CompilerOptions {
  const configFileName = ts.findConfigFile(
    "./",
    ts.sys.fileExists,
    "tsconfig.build.json"
  );
  const configFile = ts.readConfigFile(configFileName, ts.sys.readFile);
  const compilerOptions = ts.parseJsonConfigFileContent(
    configFile.config,
    ts.sys,
    "./"
  );
  return compilerOptions.options;
}

function transformImport(
  importPath: string,
  sf: ts.SourceFile,
  config: ts.CompilerOptions
): string {
  const sfParsed = path.parse(sf.fileName);
  const pathFromRoot = path.isAbsolute(sf.fileName)
    ? sfParsed.dir.replace(process.cwd() + "/", "")
    : sfParsed.dir;

  // remove the quotes surrounding the import path.
  importPath = importPath.substring(1, importPath.length - 1);
  const importPathParsed = path.parse(importPath);
  const pathParts = importPathParsed.dir.split("/");
  const importPathIsAliased =
    pathParts?.length > 0 && pathParts[0] in config.paths;

  if (!importPathIsAliased && !importPathParsed.dir.startsWith(".")) {
    return importPath;
  }

  if (importPathIsAliased) {
    // expand aliased paths.
    importPathParsed.dir = importPathParsed.dir.replace(
      pathParts[0],
      config.paths[pathParts[0]][0]
    );
    importPathParsed.dir = path.relative(pathFromRoot, importPathParsed.dir);
    if (!importPathParsed.dir) importPathParsed.dir = ".";
    else if (!importPathParsed.dir.startsWith("."))
      importPathParsed.dir = "./" + importPathParsed.dir;
  }

  // add .js to make it esm
  importPathParsed.base = importPathParsed.base + ".js";
  return path.format(importPathParsed);
}

function importVisitor(
  ctx: ts.TransformationContext,
  sf: ts.SourceFile,
  config: ts.CompilerOptions
) {
  const visitor: ts.Visitor = (node: ts.Node): ts.Node => {
    if (ts.isImportDeclaration(node) && !node.importClause?.isTypeOnly) {
      return ctx.factory.updateImportDeclaration(
        node,
        node.decorators,
        node.modifiers,
        node.importClause,
        ctx.factory.createStringLiteral(
          transformImport(node.moduleSpecifier.getText(), sf, config)
        ),
        node.assertClause
      );
    }
    return ts.visitEachChild(node, visitor, ctx);
  };
  return visitor;
}

function transform(
  config: ts.CompilerOptions
): ts.TransformerFactory<ts.SourceFile> {
  return (ctx: ts.TransformationContext): ts.Transformer<ts.SourceFile> => {
    return (sf: ts.SourceFile) =>
      ts.visitNode(sf, importVisitor(ctx, sf, config));
  };
}

function compile() {
  fs.rmSync("dist", {
    force: true,
    recursive: true,
  });
  const config = getConfig();
  const files = globSync("src/**/*.ts");
  const compilerHost = ts.createCompilerHost(config);
  const program = ts.createProgram(files, config, compilerHost);
  program.emit(undefined, undefined, undefined, undefined, {
    before: [transform(config)],
  });
}

compile();
