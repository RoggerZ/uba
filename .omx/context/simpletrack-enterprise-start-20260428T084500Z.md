# SimpleTrack Enterprise Start Context

## Task Statement

SimpleTrack is entering a formal project start phase. The user wants to first define the product/function scope, create a phased master plan, and then refine backend implementation phase by phase. A new production-grade prototype should be written in a new directory instead of modifying the existing `simpletrack/prototype/simpletrack-umami-inspired/` prototype.

## Desired Outcome

- Restore the deep-interview question path so project intake can proceed formally.
- Use the Umami research documents as evidence, especially the "对 SimpleTrack 的启发" sections.
- Use Gemini as an external advisor and save an artifact.
- Produce a scoped, phased product blueprint before backend implementation.
- Rebuild the prototype as an enterprise-grade P1/MVP working surface in a new directory.

## Known Facts

- `D:\nvm\v22.17.0\omx.cmd --version` works and reports oh-my-codex v0.14.2.
- `D:\nvm\v22.17.0\omx.cmd question --help` works.
- The bridge command also works: `D:\nvm4w\nodejs\node.exe D:\nvm\v22.17.0\node_modules\oh-my-codex\dist\cli\omx.js question --help`.
- Gemini CLI is available as version 0.39.1.
- Existing prototype directory: `simpletrack/prototype/simpletrack-umami-inspired/`.
- New prototype should be created in a separate directory.

## Evidence From Existing Docs

- Core principle: make data trustworthy before adding rich analysis.
- P0: freeze event dictionary, field dictionary, privacy rules, UTM rules, and revenue rules.
- P1/MVP: Website creation, tracking code, pageview, event, realtime, events/properties, basic filters, simple goal.
- P2: breakdown, compare, sessions, segments, funnels, journeys, basic UTM, website board.
- P3: cohorts, retention, revenue, attribution, links, pixels.
- P4: performance, replays, teams, share URL, API key.
- MVP should not default to attribution, revenue, replays, or complex teams.

## Current Hypothesis

The first production-grade prototype should focus on the P1 trust loop:

1. Create site.
2. Install tracker.
3. Listen for first pageview/event.
4. Show realtime health.
5. Show event/property evidence.
6. Define a simple goal.
7. Show settings/data dictionary.

## Unknowns / Open Questions

- What is the single most important first-value state after tracker install?
- Should P1 optimize for "first pageview is visible" or "a key business event is faithfully recorded"?
- How strict should P1 be about schema enforcement versus flexible ingestion?
- Which product persona is the primary P1 buyer/operator?

## Decision Boundary Unknowns

- Whether the new prototype should include only P1 or also preview P2 locked states.
- Whether the project should prioritize founder/dev onboarding or growth/product analytics workflows.
- Whether the backend P1 should accept unknown event properties into quarantine or reject them immediately.

## Likely Codebase Touchpoints

- `simpletrack/docs/umami/docs/`
- `simpletrack/prototype/simpletrack-umami-inspired/`
- Proposed new directory: `simpletrack/prototype/simpletrack-enterprise-mvp/`
