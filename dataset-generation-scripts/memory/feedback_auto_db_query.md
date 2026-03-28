---
name: feedback_auto_db_query
description: User wants automatic skill invocation without manual /db-query commands
type: feedback
---

**Rule:** Skills should automatically trigger based on context, not require manual invocation with `/skill-name`.

**Why:** User explicitly asked "How can I make claude code aware when to use the db-query skill... I don't want to invoke the skill manually every time"

**How to apply:**
- Add comprehensive `triggers` list to skill frontmatter with relevant keywords
- Update CLAUDE.md with "When to Use Database Queries" section
- Be proactive: if user asks about data, statistics, or verification, run queries immediately
- Don't wait for explicit `/db-query` command
- Common triggers: "how many", "count", "check", "find", "show", "verify", "missing", "database", "table"

**Example:**
- User says: "How many wards are there?"
- Response: Immediately run `SELECT COUNT(*) FROM wards_tmp` without asking
- Don't say: "Would you like me to use /db-query?"
