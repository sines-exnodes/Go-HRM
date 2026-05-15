# MOBILE-APP Platform

**Platform:** Exnodes HRM Mobile Application
**Structure:** Independent from WEB-APP (separate epic/story numbering)
**Created:** 2026-04-16

---

## Overview

This platform contains Business Analysis documentation for the Exnodes HRM mobile application. The mobile app provides employees and managers with on-the-go access to HR functions.

## Platform Scope

- Mobile-native user experience (iOS & Android)
- Core HR functions optimized for mobile interaction
- Offline-capable where applicable
- Push notification support

## Relationship to WEB-APP

This platform maintains **independent structure** from WEB-APP:
- Fresh epic/story numbering (EP-001, US-001, etc.)
- Mobile-specific user flows and interactions
- Shared backend services with WEB-APP
- Feature parity decisions documented per-epic

## Epics

| Epic ID | Name | Status | Description |
|---------|------|--------|-------------|
| EP-001 | Foundation | Approved | Authentication, core navigation, app settings |

## Conventions

- Epic folders: `EP-###-name/`
- Story folders: `US-###-name/` within each epic
- Detail Requirements: `DR-###-###-##-feature.md` in `details/` subfolder
- Figma design context extracted via `/figma-extract`

---

**Last Updated:** 2026-04-16
