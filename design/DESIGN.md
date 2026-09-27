# Ohmline

Design system of Ohmline, an AI-powered energy management platform that turns meter readings into operational decisions: what is happening, what falls outside what is expected, what to check first and why.

The interface copy is in Spanish (quoted below exactly as it appears in the UI, with an English gloss where it helps). This document is in English.

## Principles

- **Control room, not marketing.** Light, sober surfaces, thin borders instead of shadows, a single brand color. Strong color is reserved for states.
- **Numbers rule.** Every figure uses tabular numerals and the local format (1.048 kWh, +110,7%). KPIs go in a strip divided by lines, not in loose cards.
- **Color means something.** Every state and every anomaly type has a fixed color that repeats in tables, badges, charts and tiles. A state color is never used as decoration.
- **Show the evidence.** Charts always draw the baseline band and shade the anomalous window; the AI's conclusions come with the variables and rules that support them.

## Color

- `ground` for the background, `surface` for cards and tables, `line` for borders and grids, `ink` / `ink-muted` for text.
- `brand` (petrol green) for the primary action (*Run AI Analysis*), links, completed steps and the baseline line. `brand-soft` to highlight the analysis result.
- Dark panels (sidebar, AI bar, recommended action) use `ink` as background with `brand-on-dark` as the accent.

Semantic mapping (text on background):

| Meaning | Tokens |
|---|---|
| Normal | `status-ok` / `status-ok-bg` |
| Alert, medium severity | `status-alert` / `status-alert-bg` |
| Critical, high severity, real anomaly | `status-critical` / `status-critical-bg` |
| Data quality | `type-data-quality` / `type-data-quality-bg` |
| Explainable | `type-explainable` / `type-explainable-bg` |
| False positive, low severity | `type-false-positive` / `type-false-positive-bg` |

Badges always carry text; color is never the only signal.

## Typography

A single family: **Archivo** (Google Fonts, axes `wdth` 75–125 and `wght` 400–700). Page titles at 34px, weight 700 and `font-stretch: 112%`. Sections at 18px/650, body text at 15px, supporting text at 13–14px. Everything in sentence case; no all-caps text.

## Layout

- Desktop 1440px: 232px sidebar in `ink`, content with 32px padding on top and 40px on the sides, sections 22–24px apart.
- Cards with a 1px `line` border and `radius-lg`; no shadows.
- Full-width tables with a `surface-muted` header, 60px rows, numbers right-aligned.
- Touch targets of at least 44px; visible focus with a 2px `brand` outline.

## Base components

- **KPI strip:** an N-column grid inside a card, separated by vertical borders: label (label), value (kpi), subtext (caption in `ink-muted`).
- **Status badge:** 24px tall, `radius-pill`, 7px dot + caption text.
- **Type badge:** the same, without the dot.
- **Analysis stepper:** 7 steps (Lecturas, Baseline, Detección, Correlación, Eventos, Explicación, Recomendación — readings, baseline, detection, correlation, events, explanation, recommendation); 30px circle, completed in `brand` with a white check, active with a 3px `brand` border, pending with a `line` border.
- **Series chart:** 1.6px `ink` line, `chart-band` band, dashed `brand` median, anomalous window in `status-critical-bg`, start marker in `status-critical`.
- **Daily bars:** `chart-bar` in range; `brand` or `status-critical` when they exceed the threshold; baseline as a dashed `ink` line.

## Voice

The UI is in Spanish, with short sentences in the active voice. Actions say what they do ("Marcar en investigación" — mark as under investigation, "Ver anomalías" — view anomalies). Empty states invite action ("Ejecuta el análisis IA para…" — run the AI analysis to…). No emoji.

---

# Frontend implementation guide (Vue 3 + Tailwind)

## How to use this folder

- `tokens.json`: the single source of visual values. Generate the Tailwind theme extension from it (`colors`, `fontFamily`, `fontSize`, `spacing`, `borderRadius`) with the same token names (e.g. `bg-surface`, `text-ink-muted`, `text-status-critical`, `rounded-lg` = 14px). The alias `{ink-muted}` resolves to its value.
- `mockups/*.dc.html`: the approved mockups. They are the reference for layout, hierarchy, copy and states. Styles are inline: read them as a specification, do not copy the markup. They need nothing else to be understood; the `support.js` script they reference belongs to the design canvas and is not needed.
- The confidence and priority values in the mockups are provisional; in the app they come from the engine.

## Screens

| Mockup | Vue route | Notes |
|---|---|---|
| `Login.dc.html` | `/login` | Left panel in `ink` with a headline and M-109's series as an illustration (band + line). Form with the demo credentials prefilled. |
| `Main.dc.html` | `/` | Header with *Run AI Analysis*; "Análisis IA" card with the 7-step stepper; strip of 6 KPIs; daily consumption bars + "Qué atender primero" (what to attend first); status tiles for the 12 meters. States: no analysis (KPIs "—", empty list with an invitation), running (active step, button disabled "Analizando…" — analyzing), completed. |
| `Meters.dc.html` | `/meters` | Pill-style filters with counters, search by meter code, sorting by consumption / variation / status (arrow on the active header), 14-day sparkline relative to the baseline, status and anomaly badges. Empty state for a search with no results. |
| `MeterDetail.dc.html` | `/meters/:meterId` | Back to meters; dark bar with the AI verdict and a link to the investigation; 5 KPIs; hourly chart with baseline band, median, anomalous window and start of the change; 3 small charts (voltage, current, power factor); the meter's events. |
| `Anomalies.dc.html` | `/anomalies` | Table sorted by priority (rank + score bar), type, severity, confidence (Alta/Media + value), reason, action. Legend of the 4 types below. |
| `Investigation.dc.html` | `/anomalies/:id` | Left: AI explanation (source badge LLM/template), daily comparison against the baseline, variables that changed. Right: classification with priority, confidence and its breakdown; recommended action (Marcar en investigación / Resolver → status PATCH); related events; evidence (rules fired). |

## Charts (ECharts)

- Hourly series: 1.6px `ink` line; baseline band ± 3·MAD as a stacked area in `chart-band`; dashed `brand` median; `markArea` in `status-critical-bg` for the window; `markLine` in `status-critical` for the start, with a label.
- Daily bars: normal `chart-bar`; `brand` (dashboard) or `status-critical` (investigation) when they exceed the threshold; dashed `ink` `markLine` for the baseline.
- Axes at 12px in `ink-muted`, grid in `line`, no native legend (the legend is HTML above the chart, as in the mockups).

## Formatting

- `es-CO` locale: thousands with a dot, decimals with a comma (`Intl.NumberFormat('es-CO')`). Signed variations (+110,7%). Short dates: "12 sep, 14:00".
- All figures with `font-variant-numeric: tabular-nums`.
