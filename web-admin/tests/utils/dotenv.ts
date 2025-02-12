import fs from "fs";

export function updateEnvVariable(
  envFilePath: string,
  envVarName: string,
  value: string,
): void {
  let envFileContent = fs.readFileSync(envFilePath, "utf8");

  const envVarRegex = new RegExp(`^${envVarName}=.*`, "m");
  const newEnvVarLine = `${envVarName}='${value}'`;

  if (envVarRegex.test(envFileContent)) {
    // Replace the existing line
    envFileContent = envFileContent.replace(envVarRegex, newEnvVarLine);
  } else {
    // Append the new line if the variable doesn't exist
    envFileContent += `\n${newEnvVarLine}`;
  }

  fs.writeFileSync(envFilePath, envFileContent);
}
