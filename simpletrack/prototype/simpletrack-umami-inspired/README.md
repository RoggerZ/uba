# SimpleTrack Umami-Inspired Prototype

This is a static prototype used for product review before backend implementation.

## Scope

- Inspired by the logged-in Umami dashboard layout
- Reduced to a SimpleTrack MVP information architecture
- Uses mock data only

## Files

- `login.html`: login start
- `create-site.html`: create first site
- `install.html`: install snippet
- `first-data.html`: activation success bridge
- `index.html`: overview dashboard
- `events.html`: event leaderboard
- `funnels.html`: canonical funnel
- `insights.html`: weekly brief concept
- `team.html`: team access page
- `settings.html`: install and domain settings
- `styles.css`: visual system and layout
- `script.js`: small interactions for ranges, routing, copy snippet

## Review Focus

- Whether the onboarding flow is short enough
- Whether the information density feels right
- Whether `Overview / Funnels / Events / Insights / Site Settings / Team` is enough for MVP
- Whether the left navigation and top toolbar structure should stay this close to Umami

## Run

Any static server works. Example:

```bash
python -m http.server 3456
```
