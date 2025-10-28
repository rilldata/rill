#!/usr/bin/env python3
"""
Reads all inline PR review comments, generates a summary, and applies changes using Claude.
"""

import os
import sys
from pathlib import Path
from github import Github
from anthropic import Anthropic

# Environment variables
GITHUB_TOKEN = os.environ.get("GITHUB_TOKEN")
REPO = os.environ.get("REPO")
PR_NUMBER = int(os.environ.get("PR_NUMBER", 0))
ANTHROPIC_API_KEY = os.environ.get("ANTHROPIC_API_KEY")
CLAUDE_MODEL = os.environ.get("CLAUDE_MODEL", "claude-sonnet-4-5-20250929")

def main():
    if not all([GITHUB_TOKEN, REPO, PR_NUMBER, ANTHROPIC_API_KEY]):
        print("❌ Missing required environment variables")
        sys.exit(1)

    # Initialize clients
    gh = Github(GITHUB_TOKEN)
    repo = gh.get_repo(REPO)
    pr = repo.get_pull(PR_NUMBER)
    anthropic = Anthropic(api_key=ANTHROPIC_API_KEY)

    print(f"📋 Processing PR #{PR_NUMBER}: {pr.title}")

    # Get all review comments (inline comments on code)
    review_comments = list(pr.get_review_comments())
    
    # Use all review comments
    doclaude_comments = review_comments

    if not doclaude_comments:
        print("ℹ️ No review comments found in this PR")
        sys.exit(0)

    print(f"✅ Found {len(doclaude_comments)} review comment(s)")

    # Read documentation instructions
    instructions_path = Path(".claude/instructions.md")
    if instructions_path.exists():
        instructions = instructions_path.read_text()
        print(f"📖 Loaded instructions from {instructions_path}")
    else:
        instructions = "Follow standard documentation best practices."
        print("⚠️ No instructions file found, using defaults")

    # Build context from all DoClaude comments
    comments_context = []
    for i, comment in enumerate(doclaude_comments, 1):
        comments_context.append(f"""
### Comment {i}
**File**: `{comment.path}`
**Line**: {comment.position or comment.original_position or 'N/A'}
**Author**: @{comment.user.login}
**Comment**: {comment.body}

**Diff context**:
```diff
{comment.diff_hunk}
```
""")

    comments_summary = "\n".join(comments_context)

    # Build prompt for Claude
    prompt = f"""You are a documentation engineer working on a pull request.

## Documentation Guidelines
{instructions}

## PR Context
**PR #{PR_NUMBER}**: {pr.title}
{pr.body or '(No description)'}

## Inline Review Comments Requesting Changes
{comments_summary}

## Your Task
1. **Read each DoClaude comment** and understand what changes are being requested
2. **Apply the documentation guidelines** from the instructions above
3. **Generate file changes** that address ALL the DoClaude comments
4. **Output a summary** followed by the actual file changes

## Output Format

First, provide a brief summary in this format:

```summary
📝 Summary of Changes:
- [File path]: [Brief description of change]
- [File path]: [Brief description of change]
...
```

Then, for each file that needs changes, output the complete updated file content in this format:

```file:path/to/file.md
[Complete file contents with your changes applied]
```

**IMPORTANT**:
- Output COMPLETE file contents, not diffs
- Make sure all changes align with the documentation guidelines
- Address ALL DoClaude comments
- Preserve all existing formatting, frontmatter, and structure
- Do not create new files, only edit existing ones
"""

    print("\n🤖 Calling Claude API...")
    
    try:
        response = anthropic.messages.create(
            model=CLAUDE_MODEL,
            max_tokens=8000,
            messages=[{
                "role": "user",
                "content": prompt
            }]
        )
        
        claude_output = response.content[0].text
        print("\n" + "="*80)
        print("CLAUDE RESPONSE")
        print("="*80)
        print(claude_output)
        print("="*80 + "\n")

        # Parse Claude's output and apply changes
        apply_changes(claude_output)

    except Exception as e:
        print(f"❌ Error calling Claude API: {e}")
        sys.exit(1)

def apply_changes(claude_output):
    """Parse Claude's output and write file changes."""
    
    import re
    
    # Extract summary
    summary_match = re.search(r'```summary\n(.*?)\n```', claude_output, re.DOTALL)
    if summary_match:
        summary = summary_match.group(1)
        print("\n📝 SUMMARY OF CHANGES:")
        print(summary)
        print()

    # Extract file blocks: ```file:path/to/file.md ... ```
    # Split by file markers, then find the end of each block
    file_blocks = re.split(r'```file:', claude_output)
    
    changes_applied = 0
    for block in file_blocks[1:]:  # Skip the first split (before any file marker)
        # Extract file path (everything up to first newline)
        lines = block.split('\n', 1)
        if len(lines) < 2:
            continue
            
        file_path = lines[0].strip()
        remaining = lines[1]
        
        # Find the closing ``` that ends this file block
        # Look for ``` at the start of a line, potentially followed by EOF or another ```file:
        end_match = re.search(r'\n```\s*(?=\n|$)', remaining)
        if not end_match:
            print(f"⚠️ Could not find closing backticks for file: {file_path}")
            continue
        
        file_content = remaining[:end_match.start() + 1]  # Include the final newline
        
        # Verify file exists
        if not Path(file_path).exists():
            print(f"⚠️ Skipping non-existent file: {file_path}")
            continue
        
        # Write the changes
        try:
            Path(file_path).write_text(file_content)
            print(f"✅ Updated: {file_path}")
            changes_applied += 1
        except Exception as e:
            print(f"❌ Failed to update {file_path}: {e}")
    
    if changes_applied == 0:
        print("⚠️ No file changes were applied")
    else:
        print(f"\n✅ Successfully applied changes to {changes_applied} file(s)")

if __name__ == "__main__":
    main()

