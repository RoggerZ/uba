# SimpleTrack Enterprise MVP Prototype

This is the new production-oriented SimpleTrack review prototype. It is intentionally separate from `simpletrack-umami-inspired`.

The prototype uses mature frontend tooling instead of hand-rolled UI primitives:

- Vite
- React
- TypeScript
- React Router
- Ant Design
- Ant Design Charts
- Zustand

## Scope

P1 focuses on the trust loop:

- create and configure a site
- install tracker snippet
- verify the first incoming signal
- inspect realtime intake
- inspect events and property distributions
- define a simple goal
- review data dictionary and ingestion rules

Out of P1 scope:

- Team / RBAC
- Funnels / Journeys
- Revenue / Attribution
- Replays / Performance
- Boards / Share URL / API Key

## Structure

- `index.html`: Vite entry
- `src/main.tsx`: React root, Ant Design theme
- `src/App.tsx`: route definitions
- `src/components/`: shared shell and small wrappers
- `src/domain/`: typed mock data and domain contracts
- `src/store/`: local prototype state via Zustand
- `src/views/`: route-level pages
- `src/styles.css`: enterprise console overrides

## Run

```bash
npm install
npm run dev
```

Open the Vite URL shown in the terminal.
