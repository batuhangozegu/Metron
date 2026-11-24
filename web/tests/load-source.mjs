import fs from "node:fs";
import path from "node:path";
import Module, { createRequire } from "node:module";
import ts from "typescript";

const root = path.resolve(import.meta.dirname, "..");
const requireSource = createRequire(import.meta.url);

// Compile the installed app sources in memory; no generated files or extra test dependencies.
for (const extension of [".ts", ".tsx"]) {
  Module._extensions[extension] = (module, filename) => {
    const source = fs.readFileSync(filename, "utf8").replace(/(["'])@\//g, (_, quote) => quote + path.join(root, "src").replaceAll("\\", "/") + "/");
    const { outputText } = ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2020, jsx: ts.JsxEmit.ReactJSX, esModuleInterop: true } });
    module._compile(outputText, filename);
  };
}

export function loadSource(relative) { return requireSource(path.join(root, relative)); }
