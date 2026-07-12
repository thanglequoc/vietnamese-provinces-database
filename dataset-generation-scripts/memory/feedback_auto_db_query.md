---
name: feedback_auto_db_query
description: User wants automatic skill invocation without manual commands
type: feedback
---

**Rule:** Skills should automatically trigger based on context, not require manual invocation.

**Why:** User explicitly asked "How can I make claude code aware when to use the db-query skill... I don't want to invoke the skill manually every time"

**Current location:** `AGENTS.md` → `## Database Query Skill` section

**How to apply:**
- The AGENTS.md section contains explicit auto-trigger keywords and patterns
- Be proactive: if user asks about data, statistics, or verification, run queries immediately
- Don't wait for explicit commands
- Common triggers: "how many", "count", "check", "find", "show", "verify", "missing", "database", "table"

**Example:**
- User says: "How many wards are there?"
- Response: Immediately run `docker exec vn_provinces_postgres_container psql -U postgres -d vn_provinces_tmp -c "SELECT COUNT(*) FROM wards_tmp"`
- Don't say: "Would you like me to query the database?"