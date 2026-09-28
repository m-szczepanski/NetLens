# NetLens frontend

SvelteKit (Svelte 5) + TypeScript + Tailwind CSS. The UI is a static SPA built with
`@sveltejs/adapter-static`; in production the `build/` output is embedded into the Go binary,
and the frontend talks to the backend on the same origin.

## Development

```sh
npm install
npm run dev
```

The dev server runs independently of the Go backend. To point it at a local backend, set the
base URLs in `.env` (or the environment) and restart:

```sh
PUBLIC_BACKEND_HTTP_URL=http://localhost:8080 npm run dev
```

`PUBLIC_BACKEND_WS_URL` is optional; by default the WebSocket URL is derived from the HTTP URL
(`http` → `ws`, `https` → `wss`). Resolution lives in `src/lib/config.ts`.

## Checks

```sh
npm run check   # svelte-check, must be clean
npm run lint    # prettier + eslint
npm run format  # prettier --write
npm test        # vitest unit tests
npm run build   # static SPA in build/
```

To recreate this project from scratch:

```sh
npx sv@0.17.1 create --template minimal --types ts --add prettier eslint vitest="usages:unit" tailwindcss="plugins:none" sveltekit-adapter="adapter:static" frontend
```
