import { writeFile, mkdir } from 'fs/promises';
import { join, dirname } from 'path';
import { docs_v1 } from 'googleapis';

export interface MetricsData {
  currentPeriod: Record<string, number>;
  previousPeriod: Record<string, number>;
  periodLabels: {
    current: string;
    previous: string;
  };
}

/**
 * Writes report content to a local file
 */
export async function writeLocalReport(content: string, filePath: string): Promise<string> {
  try {
    const dir = dirname(filePath);

    await mkdir(dir, { recursive: true });

    await writeFile(filePath, content, 'utf8');

    return filePath;
  } catch (error) {
    throw new Error(
      `Failed to write local file: ${error instanceof Error ? error.message : 'Unknown error'}`,
    );
  }
}

/**
 * Generates a default local file path if none is provided
 */
export function getDefaultLocalPath(title: string, extension: string = 'md'): string {
  const sanitizedTitle = title
    .replace(/[^a-zA-Z0-9\s-]/g, '')
    .replace(/\s+/g, '-')
    .toLowerCase();
  const timestamp = new Date().toISOString().slice(0, 19).replace(/:/g, '-');
  return join(process.cwd(), 'reports', `${sanitizedTitle}-${timestamp}.${extension}`);
}

/**
 * Converts markdown content to Google Docs format requests
 */
export function convertMarkdownToGoogleDocsRequests(markdown: string): docs_v1.Schema$Request[] {
  const requests = [];
  const lines = markdown.split('\n');
  let insertIndex = 1;

  for (const line of lines) {
    if (line.startsWith('# ')) {
      // Header 1
      requests.push({
        insertText: {
          location: { index: insertIndex },
          text: line.substring(2) + '\n',
        },
      });
      requests.push({
        updateTextStyle: {
          range: {
            startIndex: insertIndex,
            endIndex: insertIndex + line.substring(2).length,
          },
          textStyle: {
            fontSize: { magnitude: 20, unit: 'PT' },
            bold: true,
          },
          fields: 'fontSize,bold',
        },
      });
      insertIndex += line.substring(2).length + 1;
    } else if (line.startsWith('## ')) {
      // Header 2
      requests.push({
        insertText: {
          location: { index: insertIndex },
          text: line.substring(3) + '\n',
        },
      });
      requests.push({
        updateTextStyle: {
          range: {
            startIndex: insertIndex,
            endIndex: insertIndex + line.substring(3).length,
          },
          textStyle: {
            fontSize: { magnitude: 16, unit: 'PT' },
            bold: true,
          },
          fields: 'fontSize,bold',
        },
      });
      insertIndex += line.substring(3).length + 1;
    } else {
      // Regular text
      requests.push({
        insertText: {
          location: { index: insertIndex },
          text: line + '\n',
        },
      });
      insertIndex += line.length + 1;
    }
  }

  return requests;
}

/**
 * Generates comparison report content from metrics data
 */
export function generateComparisonReportContent(
  metricsData: MetricsData,
  insights: string[],
): string {
  const { currentPeriod, previousPeriod, periodLabels } = metricsData;

  let content = `# Metrics Comparison Report\n\n`;
  content += `## Executive Summary\n\n`;
  content += `This report compares key metrics between ${periodLabels.previous} and ${periodLabels.current}.\n\n`;

  content += `## Metrics Overview\n\n`;

  // Compare each metric
  for (const [metric, currentValue] of Object.entries(currentPeriod)) {
    const previousValue = previousPeriod[metric];
    if (previousValue !== undefined) {
      const change = ((currentValue - previousValue) / previousValue) * 100;
      const direction = change > 0 ? '↑' : change < 0 ? '↓' : '→';

      content += `### ${metric}\n`;
      content += `- ${periodLabels.current}: ${currentValue}\n`;
      content += `- ${periodLabels.previous}: ${previousValue}\n`;
      content += `- Change: ${direction} ${Math.abs(change).toFixed(2)}%\n\n`;
    }
  }

  content += `## Key Insights\n\n`;
  insights.forEach((insight, index) => {
    content += `${index + 1}. ${insight}\n`;
  });

  content += `\n---\n\nReport generated on ${new Date().toLocaleDateString()}\n`;

  return content;
}
