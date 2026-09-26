# Ohmline

Sistema de diseño de Ohmline, una plataforma de gestión energética con IA que convierte lecturas de medidores en decisiones operativas: qué pasa, qué se sale de lo esperado, qué revisar primero y por qué.

## Principios

- **Sala de control, no marketing.** Superficies claras y sobrias, bordes finos en vez de sombras, un único color de marca. El color fuerte se reserva para estados.
- **Los números mandan.** Toda cifra usa números tabulares y formato local (1.048 kWh, +110,7%). Los KPIs van en una franja dividida por líneas, no en tarjetas sueltas.
- **El color significa algo.** Cada estado y cada tipo de anomalía tiene un color fijo que se repite en tablas, badges, gráficas y mosaicos. Nunca se usa un color de estado como decoración.
- **Mostrar la evidencia.** Las gráficas siempre dibujan la banda de baseline y sombrean la ventana anómala; las conclusiones de la IA van acompañadas de las variables y reglas que las sustentan.

## Color

- `ground` de fondo, `surface` para tarjetas y tablas, `line` para bordes y rejillas, `ink` / `ink-muted` para texto.
- `brand` (verde petróleo) para la acción principal (*Run AI Analysis*), enlaces, pasos completados y la línea de baseline. `brand-soft` para resaltar el resultado del análisis.
- Paneles oscuros (barra lateral, barra de IA, acción recomendada) usan `ink` de fondo con `brand-on-dark` como acento.

Mapeo semántico (texto sobre fondo):

| Significado | Tokens |
|---|---|
| Normal | `status-ok` / `status-ok-bg` |
| Alerta, severidad media | `status-alert` / `status-alert-bg` |
| Crítico, severidad alta, anomalía real | `status-critical` / `status-critical-bg` |
| Calidad de datos | `type-data-quality` / `type-data-quality-bg` |
| Explicable | `type-explainable` / `type-explainable-bg` |
| Falso positivo, severidad baja | `type-false-positive` / `type-false-positive-bg` |

Los badges siempre llevan texto; el color nunca es la única señal.

## Tipografía

Una sola familia: **Archivo** (Google Fonts, ejes `wdth` 75–125 y `wght` 400–700). Títulos de página a 34px, peso 700 y `font-stretch: 112%`. Secciones a 18px/650, texto a 15px, apoyo a 13–14px. Todo en sentence case; nada en mayúsculas sostenidas.

## Layout

- Escritorio 1440px: barra lateral de 232px en `ink`, contenido con padding 32px arriba y 40px a los lados, secciones separadas 22–24px.
- Tarjetas con borde `line` 1px y `radius-lg`; sin sombras.
- Tablas a ancho completo con cabecera `surface-muted`, filas de 60px, números alineados a la derecha.
- Objetivos táctiles de 44px mínimo; foco visible con contorno `brand` de 2px.

## Componentes base

- **Franja de KPIs:** rejilla de N columnas dentro de una tarjeta, separadas por bordes verticales: etiqueta (label), valor (kpi), subtexto (caption en `ink-muted`).
- **Badge de estado:** 24px de alto, `radius-pill`, punto de 7px + texto caption.
- **Badge de tipo:** igual, sin punto.
- **Stepper del análisis:** 7 pasos (Lecturas, Baseline, Detección, Correlación, Eventos, Explicación, Recomendación); círculo de 30px, completado en `brand` con check blanco, activo con borde `brand` de 3px, pendiente con borde `line`.
- **Gráfica de serie:** línea `ink` 1,6px, banda `chart-band`, mediana en `brand` discontinua, ventana anómala en `status-critical-bg`, marcador de inicio en `status-critical`.
- **Barras diarias:** `chart-bar` en rango; `brand` o `status-critical` cuando superan el umbral; baseline como línea `ink` discontinua.

## Voz

Español, frases cortas y en voz activa. Las acciones dicen lo que hacen ("Marcar en investigación", "Ver anomalías"). Los estados vacíos invitan a actuar ("Ejecuta el análisis IA para…"). Sin emoji.

---

# Guía de implementación para el frontend (Vue 3 + Tailwind)

## Cómo usar esta carpeta

- `tokens.json`: fuente única de valores visuales. Generar a partir de él la extensión del tema de Tailwind (`colors`, `fontFamily`, `fontSize`, `spacing`, `borderRadius`) con los mismos nombres de token (p. ej. `bg-surface`, `text-ink-muted`, `text-status-critical`, `rounded-lg` = 14px). El alias `{ink-muted}` se resuelve a su valor.
- `mockups/*.dc.html`: los mockups aprobados. Son la referencia de layout, jerarquía, textos y estados. Los estilos van en línea: leerlos como especificación, no copiar el markup. No dependen de nada para entenderse; el script `support.js` que referencian es del lienzo de diseño y no hace falta.
- Los valores de confianza y prioridad de los mockups son provisionales; en la app vienen del motor.

## Pantallas

| Mockup | Ruta Vue | Notas |
|---|---|---|
| `Login.dc.html` | `/login` | Panel izquierdo en `ink` con titular y la serie de M-109 como ilustración (banda + línea). Formulario con credenciales demo prellenadas. |
| `Main.dc.html` | `/` | Cabecera con *Run AI Analysis*; tarjeta "Análisis IA" con stepper de 7 pasos; franja de 6 KPIs; barras de consumo diario + "Qué atender primero"; mosaicos de estado de 12 medidores. Estados: sin análisis (KPIs "—", lista vacía con invitación), en curso (paso activo, botón deshabilitado "Analizando…"), completado. |
| `Meters.dc.html` | `/meters` | Filtros tipo pill con contadores, búsqueda por meter_id, orden por consumo/variación/estado (flecha en la cabecera activa), sparkline de 14 días relativa al baseline, badge de estado y de anomalía. Estado vacío de búsqueda. |
| `MeterDetail.dc.html` | `/meters/:meterId` | Volver a medidores; barra oscura con el veredicto de la IA y enlace a investigación; 5 KPIs; gráfica horaria con banda de baseline, mediana, ventana anómala e inicio del cambio; 3 gráficas pequeñas (voltaje, corriente, FP); eventos del medidor. |
| `Anomalies.dc.html` | `/anomalies` | Tabla ordenada por prioridad (rango + barra de score), tipo, severidad, confianza (Alta/Media + valor), motivo, acción. Leyenda de los 4 tipos debajo. |
| `Investigation.dc.html` | `/anomalies/:id` | Izquierda: explicación IA (badge de origen LLM/plantilla), comparación diaria contra baseline, variables que cambiaron. Derecha: clasificación con prioridad, confianza y su desglose; acción recomendada (Marcar en investigación / Resolver → PATCH estado); eventos relacionados; evidencia (reglas disparadas). |

## Gráficas (ECharts)

- Serie horaria: línea `ink` 1,6px; banda baseline ± 3·MAD como área apilada en `chart-band`; mediana en `brand` discontinua; `markArea` en `status-critical-bg` para la ventana; `markLine` en `status-critical` para el inicio, con etiqueta.
- Barras diarias: `chart-bar` normal; `brand` (resumen) o `status-critical` (investigación) cuando superan el umbral; `markLine` discontinua `ink` para el baseline.
- Ejes a 12px en `ink-muted`, rejilla en `line`, sin leyenda nativa (la leyenda va en HTML encima, como en los mockups).

## Formato

- Locale `es-CO`: miles con punto, decimales con coma (`Intl.NumberFormat('es-CO')`). Variaciones con signo (+110,7%). Fechas cortas: "12 sep, 14:00".
- Todas las cifras con `font-variant-numeric: tabular-nums`.
