---
inclusion: always
---

# WeBond Design System

Project-specific design-system steering for WeBond, created per the central policy
(`.kiro/steering/workshop-policies/design-system-policies.md`). This document holds the
**token decisions, rationale, and token→meaning map**. The **runtime source of truth** for
token values is `web/src/styles/tokens.css`; the Tailwind theme in `web/tailwind.config.js`
maps those CSS variables into utility classes. Keep all three in sync when a token changes.

Theme: **red and white, modern and friendly** (user final request, Add-on 2026-11-12c; replaces
the original white / dark-navy / navy-blue theme).

## Decisions & rationale

| Area | Decision | Rationale |
|------|----------|-----------|
| Mode | Light mode is the default (`color-scheme: light`) | Policy HARD RULE #7; dark mode only with user approval. |
| Background | Warm white page base (`--color-bg = #fffafa`), white cards | "Red and white"; the faint warm base lets white cards lift without grey. |
| Header | White, sticky, translucent (`--color-header = #ffffff`) with a bottom border | Friendly and light; the red logo tile and active nav carry the brand. |
| Body text | Warm near-black (`--color-text = #1f1416`, 17.9:1 on white) | Readable numbers/tables; warmer than pure black to suit the red palette. |
| Primary | Red (`--color-primary = #d61f2c`, white text 5.1:1) | Brand colour for primary actions, active states, links. |
| Soft tint | `--color-primary-soft = #fff1f2` (active nav, secondary buttons, highlights) | Friendly red accent without large saturated areas. Text on it: `text-primary` (4.7:1). |
| Soft tint, strong | `--color-primary-soft-strong = #ffe1e4` | Hover of soft areas. Text on it must be `text-primary-active` (7.2:1), never `text-primary` (4.2:1, fails AA). |
| Brand gradient | `--gradient-brand` red-600 → red-800 (`bg-brand`) | Auth side panel, dashboard welcome band, logo tile, summary-card top bar. Text on it is solid `text-text-inverse` only (no opacity: white/85 drops below 4.5:1). |
| Accent / focus | `--color-accent = #96101a` | Focus-visible ring, 8.8:1 on white (WCAG 2.2 AA non-text ≥ 3:1). |
| Danger / loss | Brick red `--color-danger = #b42318`, distinct from brand red | Prevents "error" reading as "brand". Loss values always keep the minus sign and down-arrow icon (never colour alone). |
| Shape | Pill buttons (`rounded-full`), radii 8/12/20/28px, soft warm shadows | Modern, friendly, touch-friendly (min 44px targets kept). |
| Chart palette | Fund = primary red, Bond = info blue, Gold = warning amber | Distinct hues; avoids red/green pairs reading as loss/gain. |
| Values | Tokens as CSS variables, consumed via Tailwind theme with `tok()` | Policy HARD RULE #1/#2. `tok()` wraps each colour in `color-mix(... <alpha-value> ...)` so opacity modifiers (`bg-text-inverse/10`) actually generate. |
| Fonts/icons | Free/open-source only (Inter + system/Noto Sans Thai fallback; SVG icons) | Policy HARD RULE #4/#3: license-free, no emoji. |

## Token → meaning map

### Surfaces
| Token | Value (light) | Meaning |
|-------|---------------|---------|
| `--color-bg` | `#fffafa` | Page base background. |
| `--color-surface` | `#ffffff` | Cards / panels / inputs. |
| `--color-surface-muted` | `#fff5f5` | Subtle section tint / hover fill. |
| `--color-border` | `#f3e3e5` | Card and control borders. |

### Text
| Token | Value | Meaning |
|-------|-------|---------|
| `--color-text` | `#1f1416` | Default body text. |
| `--color-text-muted` | `#6b6264` | Secondary / helper text (5.9:1 on white). |
| `--color-text-inverse` | `#ffffff` | Text on red / brand surfaces. |

### Brand
| Token | Value | Meaning |
|-------|-------|---------|
| `--color-header` | `#ffffff` | Header background (sticky, `bg-header/90` + blur). |
| `--color-header-text` | `#1f1416` | Header text and icons. |
| `--color-primary` | `#d61f2c` | Primary button / action / active text. |
| `--color-primary-hover` | `#b8141f` | Primary hover. |
| `--color-primary-active` | `#96101a` | Primary pressed; text on strong tint. |
| `--color-primary-soft` | `#fff1f2` | Active nav pill, secondary button, highlight. |
| `--color-primary-soft-strong` | `#ffe1e4` | Hover of soft areas. |
| `--color-accent` | `#96101a` | Focus ring. |
| `--gradient-brand` | `linear-gradient(135deg, #d61f2c, #96101a)` | Hero bands, auth panel, logo tile (`bg-brand`). |

### Semantic status
| Token | Value | Meaning |
|-------|-------|---------|
| `--color-success` | `#15803d` | Positive P/L, success states (green). |
| `--color-danger` | `#b42318` | Negative P/L, errors, destructive actions (brick red). |
| `--color-warning` | `#b45309` | Warnings / pending (amber). |
| `--color-info` | `#0369a1` | Informational notices; bond slice in charts. |

### Radius / shadow / spacing / motion
| Token | Value | Meaning |
|-------|-------|---------|
| `--radius-sm/md/lg/xl` | `8/12/20/28px` | Control, card, large-card, and hero radii. |
| `--shadow-sm/md/lg` | soft, red-tinted | Layered depth for cards/menus; `hover:shadow-md` lift on cards. |
| `--space-gutter` | `1.5rem` | Minimum horizontal page padding. |
| `--space-section` | `2.5rem` | Gap between page sections. |
| `--container-max` | `1200px` | Centered container max width. |
| `--ease-standard` / `--dur-standard` | `cubic-bezier(0.4,0,0.2,1)` / `180ms` | Standard transition; `0ms` under `prefers-reduced-motion`. |

## Usage rules (enforced)
- Components reference token-backed Tailwind classes (`bg-header`, `text-header-text`, `bg-primary`, `bg-primary-soft`, `bg-brand`, `text-success`, `rounded-lg`, `shadow-sm`, `max-w-container`, ...) — never raw hex (chart libraries read the CSS variables at runtime; their hex fallbacks mirror tokens.css).
- Status/meaning in UI pairs a semantic color with an icon and/or text label (never color alone).
- Text on `bg-brand` / `bg-primary` is solid `text-text-inverse`; `text-primary` is never used on `bg-primary-soft-strong`.
- Adding a token requires updating BOTH `web/src/styles/tokens.css` and `web/tailwind.config.js`, and this map.
- No emoji; use SVG icons. Keep light mode as default.
- Static assets that cannot read CSS variables (`web/public/favicon.svg`, `<meta name="theme-color">` in `index.html`) repeat the token hex values `#d61f2c` / `#96101a`; update them together with tokens.css.
