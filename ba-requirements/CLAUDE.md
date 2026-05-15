# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Purpose

This is a **Business Analysis documentation repository** for your project. It contains no application code—only BA documents, templates, and Claude Code automation for generating and managing requirements documentation.

**Scope boundaries:**
- ✅ Business requirements, user stories, workflows, acceptance criteria
- ✅ Process flowcharts, stakeholder analysis, business rules
- ❌ Technical implementation, database schemas, API specifications, code architecture

## Project Structure

```
docs/PLATFORMS/
├── PLATFORM-A/
│   ├── EP-001-foundation/           # Epic folder
│   │   ├── EPIC.md                  # Epic overview (must be approved first)
│   │   ├── US-001-authentication/   # User Story folder
│   │   │   ├── ANALYSIS.md          # Business analysis
│   │   │   ├── REQUIREMENTS.md      # User stories & specs
│   │   │   ├── FLOWCHART.md         # Mermaid diagrams
│   │   │   └── TODO.yaml            # BA task tracking (TK-001-XX)
│   │   └── US-002-user-management/
│   └── EP-002-core-features/        # Another Epic
├── PLATFORM-B/
├── PLATFORM-C/
└── ...
```

## Naming Convention (Jira-Compatible)

| Level | Format | Example | Scope |
|-------|--------|---------|-------|
| Epic | `EP-XXX` | `EP-001`, `EP-002` | Per platform |
| User Story | `US-XXX` | `US-001`, `US-002` | Per epic |
| Task ID | `TK-XXX-YY` | `TK-001-01` | Per story |

**Full path provides uniqueness:** `PLATFORM-A/EP-001/US-001` ≠ `PLATFORM-B/EP-001/US-001`

## Critical Workflow: EPIC-First Flow

**MANDATORY sequence when creating documentation:**

1. Create/update `EPIC.md` in Epic folder (e.g., `EP-001-foundation/EPIC.md`)
2. Set status to `approved` in metadata
3. Run `/validate-epic EP-001`
4. **Only then** create stories with `/new-story US-001-description`

Pre-flight checks enforce this—stories cannot be created until parent EPIC is approved.

## Slash Commands

| Command | Purpose | Example |
|---------|---------|---------|
| `/new-story` | Create user story folder with templates | `/new-story US-001-auth` |
| `/validate-epic` | Validate EPIC document structure | `/validate-epic EP-001` |
| `/validate-story` | Validate user story completeness | `/validate-story US-001-auth` |
| `/update-todo` | Update task status | `/update-todo US-001 "task" completed` |
| `/generate-flowchart` | Create Mermaid diagrams | `/generate-flowchart US-001 "login flow"` |
| `/requirements` | Generate formal requirements spec | `/requirements US-001` |
| `/weekly-report` | Generate progress report | `/weekly-report EP-001` |
| `/figma-extract` | Extract Figma design context | `/figma-extract US-001-auth` |
| `/figma-design-tokens` | Extract design tokens | `/figma-design-tokens` |
| `/analyze-input` | Process multi-source input | `/analyze-input all` |
| `/create-revision` | Track major document changes | `/create-revision path "reason"` |
| `/update-related` | Propagate changes to related docs | `/update-related` |
| `/sync-specs` | Sync specs/ folder with BA docs | `/sync-specs scan` |
| `/dr-agent` | Write a Detail Requirement via dr-agent | `/dr-agent US-001-request-tickets "Request Ticket Details"` |

## Figma Integration

Figma MCP server connects via `http://127.0.0.1:3845/mcp`. Before using Figma commands:
1. Open Figma Desktop
2. Switch to Dev Mode (Shift+D)
3. Enable "Desktop MCP server" in inspect panel
4. Select relevant frame/component

Design context creates `[ADD-ON]` sections in ANALYSIS.md and REQUIREMENTS.md.

## Add-on Sections

Custom sections beyond standard templates must:
- Be marked with `[ADD-ON]` suffix: `## 7. Regulatory Compliance [ADD-ON]`
- Be business-focused (no technical details)
- Be listed in document metadata `add_on_sections`
- **Never be removed** without explicit discussion

## Document Validation Standards

**Business Focus Check:** Technical keyword ratio must be < 30% of business keywords.

**Business keywords:** business, user, workflow, process, requirement, stakeholder, customer
**Technical keywords to avoid:** database, api, endpoint, schema, component, sql, json

Convert technical language to business outcomes:
- ❌ "FastAPI endpoint returns JSON"
- ✅ "Product search returns matching items"

## Key Workflow Rules

1. **Ask Before Act**: Never proceed with unclear requirements—always ask clarifying questions first
2. **Multi-Source Processing**: When multiple inputs exist, process Text → Image → Figma in order
3. **Document Linking**: All documents must have `related_documents` metadata for bidirectional linking
4. **Revision Control**: Major changes (>30% content) require `/create-revision`
5. **Conflict Resolution**: When sources conflict, halt and ask user which is authoritative

## Session Management

Context auto-saves to `.claude/sessions/YYYY-MM-DD.md` when tokens approach limit. Sessions auto-load previous context on start via SessionStart hook.

## Session Initialization

**On every new session, Claude MUST read these files for project context:**

1. **Knowledge Base** (REQUIRED): [docs/knowledge/PROJECT_KNOWLEDGE.md](docs/knowledge/PROJECT_KNOWLEDGE.md)
   - Contains confirmed patterns, rules, and decisions from completed DRs
   - Essential for maintaining consistency across requirements

2. **Platform Index** (REQUIRED): Scan [docs/PLATFORMS/](docs/PLATFORMS/) directory structure
   - Identify existing EPICs and User Stories
   - Understand current documentation scope

**Execute at session start:**
```
Read docs/knowledge/PROJECT_KNOWLEDGE.md
List docs/PLATFORMS/**/EP-*/ (EPICs)
List docs/PLATFORMS/**/US-*/ (Stories)
```

## Reference Files

- [.claude/WORKFLOW_RULES.md](.claude/WORKFLOW_RULES.md) - Complete BA workflow rules (source of truth)
- [.claude/README.md](.claude/README.md) - Configuration guide
- [.claude/templates/](.claude/templates/) - Document templates
- [.claude/commands/](.claude/commands/) - Slash command definitions
