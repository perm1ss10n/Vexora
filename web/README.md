# KONYX Control Panel (Web)

Premium-tech web console for device operations, telemetry, and command dispatch.

## Getting started

```bash
npm install
npm run dev
```

App runs on `http://localhost:5173`.

## Structure

```
src/
  app/            # Layouts, routing, auth guard
  pages/          # Public marketing + auth + protected pages
  features/       # Domain hooks (devices/telemetry/commands/settings)
  components/ui/  # shadcn/ui components (Radix-based)
  api/            # Mock data client + types
  store/          # Zustand stores (auth/settings)
  styles/         # Tailwind globals and design tokens
```
