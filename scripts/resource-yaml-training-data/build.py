#!/usr/bin/env python3
"""
Scrapes real-world Rill project resource YAML files from Rill Cloud and formats
them into text files suitable for LLM training data.
"""

import json
import os
import subprocess

SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
OUTPUT_DIR = os.path.join(SCRIPT_DIR, "output")
RESOURCE_TYPES = [
    "connector",
    "model",
    "metrics_view",
    "explore",
    "canvas",
    "theme",
]


def main():
    """Scrape all resource types from Rill Cloud and format them as text files."""
    os.makedirs(OUTPUT_DIR, exist_ok=True)

    for resource_type in RESOURCE_TYPES:
        scrape_resources(resource_type)
        format_resources(resource_type)


def scrape_resources(resource_type):
    """Dump resources of a given type from Rill Cloud to a JSON file."""
    output_path = os.path.join(OUTPUT_DIR, f"{resource_type}.json")
    with open(output_path, "w") as f:
        subprocess.run(
            [
                "rill",
                "sudo",
                "project",
                "dump-resources",
                "--include-files",
                "--type",
                resource_type,
            ],
            stdout=f,
            check=True,
        )
    print(f"Scraped {resource_type} resources to {output_path}")


def format_resources(resource_type):
    """Convert JSON dump to a readable text file with YAML content blocks."""
    json_path = os.path.join(OUTPUT_DIR, f"{resource_type}.json")
    txt_path = os.path.join(OUTPUT_DIR, f"{resource_type}.txt")

    with open(json_path) as f:
        data = json.load(f)

    lines = [f"# {resource_type.replace('_', ' ').title()} examples"]
    included_count = 0

    for item in data:
        try:
            file_path = item["meta"]["filePaths"][0]
        except (KeyError, IndexError, TypeError):
            file_path = None
        if not file_path:
            continue
        if file_path.endswith(".sql"):
            continue

        content = item.get("file_content", "")
        if not content:
            continue

        lines.append(f"## Path: {file_path}")
        if resource_type == "model":
            lines.append(f"Input connector: {item['spec']['inputConnector']}")
            lines.append(f"Output connector: {item['spec']['outputConnector']}")
        lines.append("```yaml")
        lines.append(content)
        lines.append("```")
        lines.append("")
        included_count += 1

    with open(txt_path, "w") as f:
        f.write("\n".join(lines))

    print(
        f"Formatted {included_count}/{len(data)} {resource_type} resources to {txt_path}"
    )


if __name__ == "__main__":
    main()
