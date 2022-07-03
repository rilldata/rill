export function getComponentsAndCode(components, code) {
  const exampleComponents = Object.entries(components);
  return exampleComponents.map(([exampleFile, example]) => {
    return {
      name: exampleFile,
      component: example.default,
      code: code[exampleFile],
    };
  });
}
