# AIR Dashboard: Frontend Spec

This is the hand-off doc for building the AIR dashboard UI. It covers what AIR is, what already exists, what each screen must do, and which API endpoint backs it.

---

## 1. What AIR is (read this first)

AIR is a self-hosted **over-the-air (OTA) update server** for Expo / React Native apps. Developers publish new JS bundles with a CLI, and phones running the app download them without going through the App Store or Play Store.

The dashboard is where people **see and control** those updates. It does **not** publish updates; the CLI does that.

Terms you'll see everywhere:

| Term | Meaning |
|---|---|
| **App** | One mobile app (e.g. "Acme Shop"). Identified in URLs by its `slug` (`acme-shop`). |
| **Platform** | `ios` or `android`. An app can have one or both. |
| **Channel** | A release track, e.g. `production`, `staging`. Free text chosen at publish time. |
| **Runtime version** | Compatibility label of the native binary (e.g. `1.4.0`). A device only gets updates with the **exact** same runtime version. |
| **Update** | One publish for one platform. A single `cli publish` usually creates **two** updates (iOS + Android) that share a `group_id`. |
| **Group** | All updates from one publish (same `group_id`). **The UI should treat a group as one row.** A group can contain **only one platform** (`air publish --platform ios`). |
| **Per-platform release** | iOS and Android are independent. Each platform is always on its **own** latest update, and its rollout % can differ inside the same group (e.g. iOS 100%, Android 10%). |
| **Rollout %** | Share of devices (0–100) that get this update. 100 = everyone, 0 = paused. |
| **Rollback to embedded** | A special update (`kind: "rollback_to_embedded"`) that tells devices to go back to the JS bundle shipped inside the store binary. |
| **API key** | Secret the CLI uses to publish. Belongs to one app. Shown in full **only once**, when it's created. |
| **Code signing** | Optional per app. Devices verify updates against a certificate. |

---

## 2. Tech stack and conventions (already set up, don't change)

Location: `server/web/`. Built output goes to `server/web/dist` and is embedded into the Go binary.

- **React 19 + Vite + TypeScript**
- **TanStack Router** (file-based routes in `src/routes/`, `routeTree.gen.ts` is generated)
- **TanStack Query** for all server data (`src/hooks/use-*.ts` hold `queryOptions` + mutations)
- **ky** HTTP client in `src/lib/api.ts` (prefix `/api/v1`, cookie session, throws `ApiError` with `status` and per-field `fields`)
- **shadcn/ui on Base UI** components in `src/components/ui/`
- **Tailwind v4**, tokens in `src/index.css`
- **react-hook-form** for forms, **sonner** for toasts, **date-fns** for dates, **Hugeicons** for icons
- Package manager: **bun**

Running it:

```bash
make up && make migrate-up
make create-admin email=you@example.com name="You"   # prints a password
make run        # Go API on :8080
make web-dev    # Vite on :5173, proxies /api → :8080
```

### Conventions to follow

- **Look at existing pages first** (`routes/_authed/users.tsx`, `routes/_authed/apps/$slug/members.tsx`) and copy their patterns.
- Reuse the shared building blocks: `Page`, `ErrorBanner`, `EmptyState`, `TableCard`, `Th`, `theadClass` (`components/page.tsx`); `Tag`, `StatusDot` (`components/tags.tsx`); `DeleteDialog` (`components/delete-dialog.tsx`).
- All fetching goes through a hook in `src/hooks/`. Components don't call `api` directly.
- Every list has **loading (skeleton rows), empty, and error** states.
- Form validation errors from the API come back as `ApiError.fields` (`{ field: message }`). Show them under the matching input.
- Mutations: toast on success, toast on error, invalidate the related query.
- Design language (see top of `index.css`): **plain and minimal**, light only, system font, white surfaces, gray hairline borders, black primary buttons. No gradients, no illustrations, no color except status.
- Prettier + ESLint must pass (`bun run lint`, `bun run typecheck`).

---

## 3. Who sees what (permissions)

There are three kinds of people. The API enforces all of this; the UI should **hide** actions the user can't perform, not just let them fail.

| Capability | Global admin | App admin | App developer |
|---|:-:|:-:|:-:|
| See all apps | ✅ | only theirs | only theirs |
| Create app | ✅ | – | – |
| Manage users (`/users`) | ✅ | – | – |
| View updates, change rollout, roll back | ✅ | ✅ | ✅ |
| Create API keys / revoke **own** keys | ✅ | ✅ | ✅ |
| See / revoke **everyone's** API keys | ✅ | ✅ | – |
| Add / remove members, change roles | ✅ | ✅ | – |
| App settings (platforms, code signing) | ✅ | ✅ | – |

- `GET /me` → `user.is_admin` tells you if they're a global admin.
- `GET /apps/{slug}` → `my_role` (`"admin"` | `"developer"`) tells you their role in that app.
- If `GET /apps/{slug}` returns **404**, show "App not found". Non-members get 404, not 403, by design.

---

## 4. Site map

```
/login                          public
/                               → redirects to /apps
/apps                           app list (+ "New app" for global admins)
/apps/:slug                     Updates tab (default)          ← main screen
/apps/:slug/updates/:groupId    Update detail                  NEW
/apps/:slug/members             Members tab
/apps/:slug/api-keys            API keys tab
/apps/:slug/settings            Settings tab (app admin only)  NEW
/users                          user management (global admin only)
/profile                        name + change password
```

Layout: left **sidebar** (logo, Apps, Users [admins], list of the user's apps, footer with profile / logout), and a centered content column (`max-w-5xl`). The app pages have a header (← Apps, app name, role tag) and a **tab bar**: `Updates · Members · API keys · Settings`.

---

## 5. Current status

| Screen | Status |
|---|---|
| Login, auth guard, logout | ✅ done |
| Sidebar + layout | ✅ done |
| Apps list + create app dialog | ✅ done |
| App layout (header + tabs) | ✅ done |
| Members tab | ✅ done |
| API keys tab (create, one-time key display, revoke) | ✅ done |
| Users page | ✅ done |
| Profile page | ✅ done |
| **Updates tab** | ⚠️ basic table only, one row per platform, no actions |
| **Update detail** | ❌ to build |
| **Rollout control** | ❌ to build |
| **Rollback** | ❌ to build |
| **Settings tab** | ❌ to build |
| **Apps list: "latest update" info** | ❌ nice to have |

**Most of the remaining work is the Updates area.** It's the reason the dashboard exists.

> ⚠️ Several backend handlers are still stubs and return **501 Not Implemented** (see §8). Build the UI against the contracts below, using mock data where needed, and handle 501 gracefully (the Updates tab already does this).

---

## 6. Screens to build

### 6.1 Updates tab: `/apps/:slug` (main screen)

**Purpose:** answer "what is live right now, for whom, and what was before it?" and let people act on it.

**Data:** `GET /api/v1/apps/{slug}/updates` → `{ updates: Update[] }` (newest first).

```ts
interface Update {
  id: string
  group_id: string
  channel: string
  platform: "ios" | "android"
  runtime_version: string
  kind: "update" | "rollback_to_embedded"
  message: string
  git_commit: string
  rollout_percent: number   // 0–100
  created_at: string        // ISO
}
```

**Layout (top to bottom):**

1. **Filter bar**
   - Channel select (options = distinct `channel` values from the data, default `production` if it exists, else first)
   - Runtime version select (distinct values for that channel, default = newest)
   - Platform segmented control: `All · iOS · Android`
   - Keep filters in the URL as search params (`?channel=production&runtime=1.4.0&platform=ios`) using TanStack Router `validateSearch`, so links can be shared.

2. **"Live now" card**, one per platform for the selected channel + runtime, **side by side**
   - The newest non-rolled-back update for that (channel, runtime, platform) is what devices get. **The two cards can show different updates**: iOS may be on "New checkout" while Android is still on "Fix back button". That's normal; don't treat it as an error.
   - Shows: message, short id, git commit (first 7 chars, monospace), published time ("3h ago", exact time in tooltip), rollout % with a progress bar.
   - If it's a `rollback_to_embedded`: red `Tag` "Rolled back to embedded bundle".
   - If rollout < 100: note like "40% of devices. The rest stay on the previous update."
   - Actions: **Change rollout**, **Roll back** (see 6.3, 6.4).

3. **History table**, **grouped by `group_id`** (one row per publish, not per platform)

   | Column | Content |
   |---|---|
   | Update | message (or "Untitled update"); below: short group id · git commit, monospace, muted |
   | Platforms | small tags `iOS` `Android`. Only the platforms in this group (single-platform publishes show one tag) |
   | Channel | text |
   | Runtime | monospace. If the platforms differ, show both (`iOS 1.2.0 · Android 1.1.0`) |
   | Rollout | `StatusDot`: green 100%, amber 1–99%, gray 0%. If the platforms differ, show one per platform (`iOS 100% · Android 10%`) |
   | Published | relative time, exact in tooltip |
   | ⋯ | menu: View details, Change rollout, Roll back to this |

   - Row click → Update detail.
   - A row that is currently live gets a "Live" tag **per platform** (`Live on iOS`). A group can be live on iOS but superseded on Android.
   - When the Platform filter is `iOS` or `Android`, only show groups that contain that platform.
   - Client-side pagination or "Show more" after 50 rows is fine for now.

**States:**
- Loading → skeleton cards + skeleton rows.
- No updates at all → `EmptyState` with a short how-to:
  ```
  No updates yet. Publish one from your app folder:
    npx air publish --channel production
  ```
  (Put the command in a copyable code block. Link to the API keys tab: "You'll need an API key." Add a one-line hint: "Add `--platform ios` or `--platform android` to publish to one platform only.")
- No updates for the selected filters → "Nothing published to *staging* for runtime *1.3.0*."
- 501 → the existing "will appear once publishing is available" message.

### 6.2 Update detail: `/apps/:slug/updates/:groupId`

Built from the same `updates` query (filter by `group_id`), so no new endpoint is needed.

- Header: message, `Live` / `Superseded` / `Rolled back` tag, published time.
- Metadata list: group id (copy button), channel, runtime version, git commit (copy button), kind.
- Per-platform section (iOS / Android), **only for the platforms in the group**: update id (copy), runtime version, rollout %, status (`Live` / `Superseded` / `Rolled back`), and a **Change rollout** button scoped to that platform.
- Group actions: Change rollout (all platforms), Roll back to this update.
- "← Updates" back link that keeps the previous filters.

### 6.3 Change rollout (dialog)

- Opened from the Live card, a history row, or the detail page.
- **Platform** segmented control: `All platforms · iOS · Android`. Only show the platforms in the group, and hide the control entirely for single-platform groups.
  - Opened from a Live card or a per-platform section → preselect that platform.
  - Opened from a history row → preselect `All platforms`.
  - Show the current value for each platform above the slider (`iOS 100% · Android 10%`).
- Slider **plus** a number input (0–100), with preset buttons `10% · 25% · 50% · 100%`. When `All platforms` is selected and the current values differ, start the slider at the lowest one and note "This sets both platforms to the same value."
- Helper text: "Devices are bucketed consistently. A device that got this update keeps getting it as you raise the percentage."
- `0%` shows a warning: "This pauses the update. Devices that haven't downloaded it stay on the previous one."
- Submit → `PATCH /api/v1/apps/{slug}/updates/{groupId}` with `{ "rollout_percent": 40, "platform": "android" }`. Leave out `platform` for `All platforms`.
- On success: toast, invalidate the updates query.

### 6.4 Roll back (dialog)

Rolling back is scary, so it gets a clear confirmation dialog, and the destructive button is red.

Two options (radio):
1. **Roll back to a previous update**: pick one from a list of earlier updates in the same channel + runtime (default = the one right before the live one). Devices get that bundle again.
2. **Roll back to the embedded bundle**: devices drop all OTA updates and run the JS that shipped inside the store build.

Also: platform choice (`Both / iOS / Android`), prefilled from the context. Rolling back one platform never touches the other. The "previous update" list only shows updates that contain the selected platform(s).

Summary line before the button, e.g.: *"Devices on **production**, runtime **1.4.0**, **iOS + Android** will switch to 'Fix checkout crash' (a1b2c3d) on their next launch."*

Submit → `POST /api/v1/apps/{slug}/updates/rollback` (payload proposed in §8, confirm it with the backend before wiring it up).

### 6.5 Settings tab: `/apps/:slug/settings` (app admins only)

Hide the tab for developers. If a developer opens the URL directly, show "Only app admins can change settings."

Sections, each a bordered card with a title and short description:

1. **General**: app name, slug (read only, with copy). *(Rename/delete need backend endpoints that don't exist yet. Leave these out or disable them.)*
2. **Platforms**
   - A row for iOS and one for Android: bundle id / package name, an **Enabled** switch.
   - "Add platform" if one is missing → dialog: platform + bundle id → `POST /apps/{slug}/platforms` `{ platform, bundle_id }`.
   - Toggle → `PATCH /apps/{slug}/platforms/{platform}` `{ enabled }`. Turning one **off** needs a confirm dialog: "Devices on Android will stop receiving updates."
3. **Code signing**
   - Status: `Enabled` (key id) / `Not configured`.
   - "Configure" → dialog with a textarea for the PEM certificate, plus the key id (default `main`). Explain that the signed no-update directive and signature come from the CLI (`air signing setup`), and give a textarea for each. → `PUT /apps/{slug}/code-signing`.
4. **Client setup (help)**: a read-only snippet showing what to put in the Expo `app.json`:
   ```json
   "updates": {
     "url": "https://<this-server>/api/manifest?app=<slug>",
     "requestHeaders": { "expo-channel-name": "production" }
   },
   "runtimeVersion": "1.0.0"
   ```
   Fill in the server origin from `window.location.origin`, with a copy button.

### 6.6 Apps list improvements (nice to have)

Each app card/row: name, slug, your role, platform icons, **last published** ("2d ago"). This needs extra fields from `GET /apps` (see §8). Skip it if the backend doesn't add them.

---

## 7. Shared UX rules

- **Copy buttons** on every id, key, commit, and code snippet (icon button, toast "Copied").
- **Ids**: show 8 chars in monospace, full value in a tooltip.
- **Times**: relative (`formatDistanceToNow`), with the exact local time in a tooltip.
- **Destructive actions** (revoke key, remove member, rollback, disable platform, 0% rollout) always go through a confirm dialog that says exactly what will happen.
- **Buttons** show a spinner and are disabled while their mutation is pending.
- **Keyboard**: dialogs close with Esc, and forms submit with Enter.
- **Responsive**: works down to ~768px (sidebar collapses to a sheet; that's already built). Tables scroll horizontally inside their card rather than breaking the layout.
- **Accessibility**: every icon-only button has an `aria-label`, and status is never shown by color alone (always a text label too).

---

## 8. API reference and gaps

All routes are under `/api/v1`, use the session cookie, and return JSON. Errors are `{ "error": "message" }` or `{ "error": { "field": "message" } }`.

### Ready

| Method | Path | Returns |
|---|---|---|
| POST | `/auth/login` `{email,password}` | sets cookie |
| POST | `/auth/logout` | – |
| GET / PATCH | `/me` | `{ user }` |
| POST | `/me/password` | – |
| GET | `/apps` | `{ apps }` |
| POST | `/apps` `{slug,name}` | `{ app }` (global admin) |
| GET | `/apps/{slug}` | `{ app, my_role }` |
| GET / POST | `/apps/{slug}/members` | `{ members }` |
| PATCH / DELETE | `/apps/{slug}/members/{userId}` | – |
| GET / POST | `/apps/{slug}/api-keys` | `{ api_keys }` / `{ api_key, key }` |
| DELETE | `/apps/{slug}/api-keys/{keyId}` | 204 |
| GET / POST | `/users` | global admin |
| PATCH | `/users/{userId}` | global admin |

### Route exists, handler is a stub (returns 501)

| Method | Path | Body | Notes |
|---|---|---|---|
| GET | `/apps/{slug}/updates` | – | → `{ updates: Update[] }` |
| PATCH | `/apps/{slug}/updates/{groupId}` | `{ rollout_percent, platform?: "ios" \| "android" }` | no `platform` = every platform in the group. Returns `{ group_id, updates }` |
| POST | `/apps/{slug}/updates/rollback` | `{ channel, runtime_version, platforms: ["ios"] \| ["android"] \| ["ios","android"], to: "previous" \| "embedded", group_id?: string }` | see `ROADMAP.md` M3 §3.2 |
| POST | `/apps/{slug}/platforms` | `{ platform, bundle_id }` | |
| PATCH | `/apps/{slug}/platforms/{platform}` | `{ enabled }` | |
| PUT | `/apps/{slug}/code-signing` | `{ certificate, key_id }` | |

### Missing (backend needs to add these for the full UI)

- `GET /apps/{slug}` should also return **`platforms`** (`[{platform, bundle_id, enabled}]`) and **`code_signing`** (`{enabled, key_id}`). The Settings tab needs them.
- `GET /apps` could include `platforms` and `last_published_at` for the apps list (§6.6).
- Rename app / delete app endpoints (Settings → General).

Until these exist, put mock data in the hook (clearly marked `// TODO: mock until backend ships`) so the UI can be built and reviewed.

---

## 9. Out of scope (for now)

- Publishing from the browser (the CLI does this)
- Device analytics / download counts / crash rates (no data is collected yet)
- Dark mode
- Audit log
- Multi-org / billing

---

## 10. Definition of done

- [ ] Updates tab: filters (URL-synced), Live cards per platform, grouped history table, all states
- [ ] Update detail page
- [ ] Change-rollout dialog wired to the PATCH endpoint, including a per-platform option
- [ ] Rollback dialog (both modes, per platform) wired to the rollback endpoint
- [ ] Checked with a **single-platform** group and with a group whose iOS and Android rollouts differ
- [ ] Settings tab: platforms, code signing, client-setup snippet; hidden for developers
- [ ] Every action hidden when the user lacks permission (§3)
- [ ] Loading / empty / error states on every data view
- [ ] No new UI libraries without asking; reuse `components/ui` and `components/page.tsx`
- [ ] `bun run lint`, `bun run typecheck`, and `bun run build` pass
- [ ] Checked by hand as global admin, app admin, and developer
