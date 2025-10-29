#!/usr/bin/env python3
"""
Reads all inline PR review comments, generates a summary, and applies changes using Claude.
"""

import os
import sys 
import re
import requests
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
    print("🔍 Fetching inline review comments from PR...")
    
    # Check reviews first
    reviews = list(pr.get_reviews())
    print(f"📊 Found {len(reviews)} review(s) on this PR:")
    for review in reviews:
        print(f"  - Review by {review.user.login}: {review.state} ({review.submitted_at})") 
    
    # Get review comments (inline comments on code lines)
    review_comments = list(pr.get_review_comments())
    
    # Debug: Show what we found
    print(f"\n📊 Found {len(review_comments)} inline review comment(s):")
    for i, comment in enumerate(review_comments, 1):
        print(f"  {i}. {comment.path}:{comment.line or comment.original_line or 'N/A'} - {comment.body[:60]}...")
    
    # Use all review comments
    doclaude_comments = review_comments

    if not doclaude_comments:
        print("⚠️ No inline review comments found in this PR")
        print("ℹ️ Note: This fetches inline code review comments only, not general PR comments")
        sys.exit(0)

    print(f"\n✅ Processing {len(doclaude_comments)} inline review comment(s)") 

    # Read documentation instructions
    instructions_path = Path(".github/scripts/instructions.md")
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
        
        # Resolve the review comments that were addressed
        resolve_review_comments(doclaude_comments, pr)
        
        # Update instructions.md based on review comments
        update_instructions_from_comments(doclaude_comments, anthropic)

    except Exception as e:
        print(f"❌ Error calling Claude API: {e}")
        sys.exit(1)

def resolve_review_comments(comments, pr):
    """Mark review comments as resolved after addressing them."""
    
    print("\n📌 Resolving review comments...")
    
    # GitHub's REST API via PyGithub doesn't support resolving comments directly
    # We need to use GraphQL API for that
    # For now, we'll reply to each comment to indicate it was addressed
    
    resolved_count = 0
    for comment in comments:
        try:
            # Reply to the comment
            comment.create_reply("✅ Addressed by DoClaude automation")
            resolved_count += 1
            print(f"✅ Marked comment on {comment.path} as addressed")
        except Exception as e:
            print(f"⚠️ Could not reply to comment {comment.id}: {e}")
    
    if resolved_count > 0:
        print(f"\n✅ Replied to {resolved_count} review comment(s)")
    
    # Also try to resolve via GraphQL if we have the necessary access
    try:
        resolve_comments_graphql(comments)
    except Exception as e:
        print(f"ℹ️ Could not resolve comments via GraphQL (this is optional): {e}")

def resolve_comments_graphql(comments):
    """Use GitHub GraphQL API to resolve review threads."""
    
    headers = {
        "Authorization": f"Bearer {GITHUB_TOKEN}",
        "Content-Type": "application/json"
    }
    
    # Group comments by pull request review thread
    # We need to get the thread ID for each comment
    resolved_threads = set()
    
    for comment in comments:
        # Get the comment's node_id
        comment_node_id = comment.raw_data.get('node_id')
        if not comment_node_id:
            continue
        
        # First, query to get the pull request review thread for this comment
        query_thread = """
        query($commentId: ID!) {
          node(id: $commentId) {
            ... on PullRequestReviewComment {
              pullRequestReview {
                id
              }
              pullRequest {
                reviewThreads(first: 100) {
                  nodes {
                    id
                    isResolved
                    comments(first: 1) {
                      nodes {
                        id
                      }
                    }
                  }
                }
              }
            }
          }
        }
        """
        
        response = requests.post(
            "https://api.github.com/graphql",
            headers=headers,
            json={"query": query_thread, "variables": {"commentId": comment_node_id}}
        )
        
        if response.status_code != 200:
            print(f"⚠️ Could not fetch thread for comment {comment.id}")
            continue
        
        data = response.json()
        
        # Find the thread containing this comment
        review_threads = data.get('data', {}).get('node', {}).get('pullRequest', {}).get('reviewThreads', {}).get('nodes', [])
        
        thread_id = None
        for thread in review_threads:
            if thread.get('isResolved'):
                continue  # Skip already resolved threads
            thread_comments = thread.get('comments', {}).get('nodes', [])
            if any(tc.get('id') == comment_node_id for tc in thread_comments):
                thread_id = thread.get('id')
                break
        
        if not thread_id or thread_id in resolved_threads:
            continue
        
        # Now resolve the thread
        resolve_mutation = """
        mutation($threadId: ID!) {
          resolveReviewThread(input: {threadId: $threadId}) {
            thread {
              id
              isResolved
            }
          }
        }
        """
        
        response = requests.post(
            "https://api.github.com/graphql",
            headers=headers,
            json={"query": resolve_mutation, "variables": {"threadId": thread_id}}
        )
        
        if response.status_code == 200:
            result = response.json()
            if not result.get('errors'):
                resolved_threads.add(thread_id)
                print(f"✅ Resolved thread for comment on {comment.path}")
        else:
            print(f"⚠️ GraphQL resolve failed: {response.text}")
    
    if len(resolved_threads) > 0:
        print(f"\n✅ Resolved {len(resolved_threads)} review thread(s) via GraphQL")

def apply_changes(claude_output):
    """Parse Claude's output and write file changes."""
    
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
        
        # Find the LAST closing ``` that ends this file block
        # We need to find all occurrences and take the last one
        matches = list(re.finditer(r'\n```\s*$', remaining, re.MULTILINE))
        if not matches:
            print(f"⚠️ Could not find closing backticks for file: {file_path}")
            continue
        
        # Use the last match (the actual closing backticks of the file block)
        end_match = matches[-1]
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

def update_instructions_from_comments(comments, anthropic):
    """Analyze review comments and update instructions.md with learnings."""
    
    instructions_path = Path(".github/scripts/instructions.md")
    if not instructions_path.exists():
        print("⚠️ Instructions file not found, skipping instruction updates")
        return
    
    # Build context from comments
    comments_text = "\n\n".join([
        f"**File**: {c.path}\n**Comment**: {c.body}" 
        for c in comments
    ])
    
    current_instructions = instructions_path.read_text()
    
    prompt = f"""You are improving documentation guidelines based on review feedback.

## Current Instructions File
{current_instructions}

## Review Comments Received
{comments_text}

## Your Task
Analyze these review comments and determine if there are patterns, rules, or guidance that should be added to the instructions file to prevent similar issues in the future.

If the comments reveal:
- New terminology standards (e.g., deprecated terms to avoid)
- Common mistakes or patterns to fix
- Style preferences
- Technical requirements
- Structural guidelines

Then output an updated version of the instructions file with a new section capturing these learnings.

**Important:**
- Only add NEW guidance that isn't already covered
- Keep all existing content
- Add a dated comment showing when/why this was added (e.g., "<!-- Added from PR #8166 review - 2024-10-29 -->")
- If no meaningful updates are needed, output "NO_UPDATES_NEEDED"

Output format:
```instructions:.github/scripts/instructions.md
[Complete updated instructions file if updates needed, or just "NO_UPDATES_NEEDED"]
```
"""

    print("\n🤖 Analyzing review comments to improve instructions...")
    
    try:
        response = anthropic.messages.create(
            model=CLAUDE_MODEL,
            max_tokens=4000,
            messages=[{
                "role": "user",
                "content": prompt
            }]
        )
        
        output = response.content[0].text
        
        # Check if updates are needed
        if "NO_UPDATES_NEEDED" in output:
            print("ℹ️ No new patterns found, instructions remain unchanged")
            return
        
        # Extract updated instructions
        # Use same approach as file extraction - find last closing backticks
        blocks = re.split(r'```instructions:', output)
        if len(blocks) < 2:
            print("⚠️ Could not parse instruction updates")
            return
        
        block = blocks[1]
        lines = block.split('\n', 1)
        if len(lines) < 2:
            print("⚠️ Could not parse instruction updates")
            return
        
        remaining = lines[1]
        matches = list(re.finditer(r'\n```\s*$', remaining, re.MULTILINE))
        if not matches:
            print("⚠️ Could not find closing backticks for instructions")
            return
        
        end_match = matches[-1]
        updated_instructions = remaining[:end_match.start() + 1]
        
        # Write updated instructions
        instructions_path.write_text(updated_instructions)
        print("✅ Updated instructions.md with learnings from review comments")
        
    except Exception as e:
        print(f"⚠️ Could not update instructions: {e}")
        # Non-fatal - continue with the main task

if __name__ == "__main__":
    main()

