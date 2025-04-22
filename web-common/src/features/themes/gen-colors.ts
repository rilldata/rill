import fs from "fs";
import path from "path";
import { fileURLToPath } from "url";
import { allColors } from "./colors.ts";
import { createDarkVariation } from "./actions.ts";
import { TailwindColorSpacing } from "./color-config.ts";
import { exec } from "child_process";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const outputPath = path.join(__dirname, "../../colors.css");

const header = `/** 
 * This file is auto-generated. 
 * Do not edit manually. 
 * Source: /web-common/src/features/themes/gen-colors.ts
 * Script: npm run gen:colors
**/\n\n`;

function generateCSSBlock(): string {
  let variables = "";
  let lightAssignments = "  /* LIGHT MODE ASSIGNMENT */\n";
  let darkAssignments = "  /* DARK MODE ASSIGNMENT */\n";

  for (const [colorName, colorMap] of Object.entries(allColors)) {
    const colorList = Object.values(colorMap);
    const darkVariants = createDarkVariation(colorName, [...colorList]);

    // Light and dark variables for each color and shade
    const lightVars = colorList
      .map(
        (color, i) =>
          `  --color-${colorName}-light-${TailwindColorSpacing[i]}: ${color.css("oklch")};`,
      )
      .join("\n");

    const darkVars = darkVariants
      .map(
        (color, i) =>
          `  --color-${colorName}-dark-${TailwindColorSpacing[i]}: ${color.css("oklch")};`,
      )
      .join("\n");

    variables += `  /* ${colorName.toUpperCase()} */\n${lightVars}\n\n${darkVars}\n\n`;

    // Assigning light and dark variables to the main color variables
    lightAssignments += `  /* ${colorName.toUpperCase()} */\n`;
    darkAssignments += `  /* ${colorName.toUpperCase()} */\n`;

    for (const key of Object.keys(colorMap)) {
      lightAssignments += `  --color-${colorName}-${key}: var(--color-${colorName}-light-${key});\n`;
      darkAssignments += `  --color-${colorName}-${key}: var(--color-${colorName}-dark-${key});\n`;
    }

    lightAssignments += "\n";
    darkAssignments += "\n";
  }

  return `:root {\n${variables}${lightAssignments}}\n\n:root.dark {\n${darkAssignments}}`;
}

const cssContent = header + generateCSSBlock();
fs.writeFileSync(outputPath, cssContent);

exec(`npx prettier --write ${outputPath}`, (error, stdout, stderr) => {
  if (error || stderr) {
    console.error(`Error running Prettier: ${error?.message || stderr}`);
    return;
  }

  console.log(`Prettier ran successfully: ${stdout}`);
});

console.log(`CSS file generated at ${outputPath}`);
