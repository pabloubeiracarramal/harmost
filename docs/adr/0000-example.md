---
id: {{sequence_number}}
title: {{short_descriptive_title}}
date: {{YYYY-MM-DD}}
status: {{Proposed | Accepted | Deprecated | Superseded}}
tags: [{{e.g., architecture, frontend, golang, database}}]
---

# ADR {{sequence_number}}: {{short_descriptive_title}}

## 1. Context and Problem Statement
*What is the problem we are trying to solve?* 
Keep this to 1-3 paragraphs. Explain the technical or business constraints that led to this discussion. Mention what we are currently doing and why it is no longer working.

## 2. Considered Options
*What were the alternatives?*
* Option A: [Brief description]
* Option B: [Brief description]
* Option C: [Brief description]

## 3. Decision
*What did we decide to do?*
State the decision clearly and directly. 

**Justification:** Explain *why* this option was chosen over the others. Was it a trade-off for performance? Developer velocity? Ecosystem compatibility? 

## 4. Consequences
*What happens now that we've made this decision?*

**Positive:**
* [e.g., Build times will decrease by 40%]
* [e.g., Frontend and backend teams can share types natively]

**Negative / Trade-offs:**
* [e.g., New developers will need to learn Go Workspaces]
* [e.g., We can no longer use tool X]

---

## 5. AI Directives (System Rules)
*Instructions for AI Agents (Claude, GitHub Copilot, Windsurf, etc.) reading this document.*

* **MUST:** [e.g., "Always use `go work use` to link new Go apps."]
* **MUST NOT:** [e.g., "Never place a `go.mod` file at the root of the monorepo."]
* **REFERENCE:** [e.g., "See `/apps/agent/README.md` for specific implementation details."]