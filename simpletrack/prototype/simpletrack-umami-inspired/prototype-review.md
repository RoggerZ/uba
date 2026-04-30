# SimpleTrack Prototype Review Notes

## Current Pages

- `login.html`
- `create-site.html`
- `install.html`
- `first-data.html`
- `index.html`
- `events.html`
- `funnels.html`
- `insights.html`
- `team.html`
- `settings.html`

## Review Questions

1. Is the page set small enough for MVP?
2. Does the left navigation feel too close to Umami, or is it an acceptable structural reference?
3. Is the `Insights` page useful enough to justify a dedicated navigation item in v1?
4. Does `Team` need its own page at MVP, or can it live inside settings?
5. Is the main `Overview` page too dense or still readable at first glance?
6. Should `Funnels` remain one canonical funnel page before opening saved funnels and comparisons?
7. Does the onboarding path feel short enough to get someone to first value?
8. Is `first-data.html` useful as a bridge page, or should users jump directly into the dashboard?

## Backend Contract Impact

If this prototype direction is accepted, backend work can be scoped around:

- overview summary endpoint
- event leaderboard endpoint
- funnel summary endpoint
- insight summary endpoint
- site settings endpoint
- team member list endpoint
