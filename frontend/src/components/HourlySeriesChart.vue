<script setup lang="ts">
import { computed } from 'vue'
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import { LineChart } from 'echarts/charts'
import {
  GridComponent,
  MarkAreaComponent,
  MarkLineComponent,
  TooltipComponent,
} from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import type { ReadingPoint } from '@/types/api'
import { cssColor } from '@/utils/theme'
import { formatDataTime, formatDecimal } from '@/utils/format'

use([
  LineChart,
  GridComponent,
  MarkAreaComponent,
  MarkLineComponent,
  TooltipComponent,
  CanvasRenderer,
])

type Metric = 'consumption' | 'voltage' | 'current' | 'pf'

const props = withDefaults(
  defineProps<{
    points: ReadingPoint[]
    metric: Metric
    /** The anomalous window to shade, if any. */
    window?: { start: string; end: string } | null
    /** Large chart: baseline band, start-of-change label and daily ticks. */
    large?: boolean
    markerLabel?: string
    height?: number
    description: string
  }>(),
  { window: null, large: false, markerLabel: '', height: 150 },
)

const METRICS = {
  consumption: {
    value: (p: ReadingPoint) => p.consumption_kwh,
    base: (p: ReadingPoint) => p.baseline_kwh,
    unit: 'kWh',
    digits: 1,
  },
  voltage: {
    value: (p: ReadingPoint) => p.voltage_v,
    base: (p: ReadingPoint) => p.baseline_voltage_v,
    unit: 'V',
    digits: 1,
  },
  current: {
    value: (p: ReadingPoint) => p.current_a,
    base: (p: ReadingPoint) => p.baseline_current_a,
    unit: 'A',
    digits: 0,
  },
  pf: {
    value: (p: ReadingPoint) => p.power_factor,
    base: (p: ReadingPoint) => p.baseline_power_factor,
    unit: '',
    digits: 2,
  },
} as const

const option = computed(() => {
  const m = METRICS[props.metric]
  const ink = cssColor('ink', '#15201C')
  const muted = cssColor('ink-muted', '#55615C')
  const line = cssColor('line', '#D8DFDB')
  const brand = cssColor('brand', '#0E6B63')
  const band = cssColor('chart-band', '#CFE3E0')
  const critical = cssColor('status-critical', '#B42318')
  const criticalBg = cssColor('status-critical-bg', '#FDECEA')

  const times = props.points.map((p) => p.timestamp)
  // One tick per day on the large chart, one every three days on the small ones.
  const step = props.large ? 24 : 72

  const lineSeries = [
    {
      name: 'baseline',
      type: 'line',
      data: props.points.map(m.base),
      showSymbol: false,
      silent: true,
      lineStyle: { color: brand, width: 1.2, type: [4, 3] },
      z: 3,
    },
    {
      name: 'value',
      type: 'line',
      data: props.points.map(m.value),
      showSymbol: false,
      lineStyle: { color: ink, width: 1.6 },
      z: 4,
      markArea: props.window
        ? {
            silent: true,
            itemStyle: { color: criticalBg },
            data: [[{ xAxis: props.window.start }, { xAxis: props.window.end }]],
          }
        : undefined,
      markLine: props.window
        ? {
            silent: true,
            symbol: 'none',
            lineStyle: { color: critical, width: 1.5, type: 'solid' },
            label: props.markerLabel
              ? {
                  show: true,
                  formatter: props.markerLabel,
                  color: critical,
                  fontSize: 13,
                  fontWeight: 600,
                  position: 'end',
                  rotate: 0,
                  align: 'right',
                  offset: [-6, 14],
                }
              : { show: false },
            data: [{ xAxis: props.window.start }],
          }
        : undefined,
    },
  ]

  // The expected range is drawn as a stacked area: an invisible base plus the band on top of it.
  const bandSeries = props.large
    ? [
        {
          name: 'band-low',
          type: 'line',
          stack: 'band',
          data: props.points.map((p) => p.band_low_kwh),
          showSymbol: false,
          silent: true,
          lineStyle: { opacity: 0 },
          tooltip: { show: false },
        },
        {
          name: 'band',
          type: 'line',
          stack: 'band',
          data: props.points.map((p) => p.band_high_kwh - p.band_low_kwh),
          showSymbol: false,
          silent: true,
          lineStyle: { opacity: 0 },
          areaStyle: { color: band, opacity: 0.9 },
          tooltip: { show: false },
        },
      ]
    : []

  return {
    animation: false,
    grid: { left: props.large ? 44 : 36, right: 8, top: props.large ? 22 : 8, bottom: 24 },
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'line', lineStyle: { color: line } },
      formatter: (params: { dataIndex: number }[]) => {
        const p = props.points[params[0].dataIndex]
        const unit = m.unit ? ` ${m.unit}` : ''
        return `${formatDataTime(p.timestamp)}<br/><b>${formatDecimal(m.value(p), (m.digits || 1) as 1 | 2)}${unit}</b><br/>Esperado ${formatDecimal(m.base(p), (m.digits || 1) as 1 | 2)}${unit}`
      },
    },
    xAxis: {
      type: 'category',
      data: times,
      boundaryGap: false,
      axisLine: { show: false },
      axisTick: { show: false },
      axisLabel: {
        color: muted,
        fontSize: 12,
        interval: (index: number) => index % step === 0,
        formatter: (value: string, index: number) => {
          const d = new Date(value)
          return index === 0 ? `${d.getUTCDate()} sep` : String(d.getUTCDate())
        },
      },
    },
    yAxis: {
      type: 'value',
      scale: props.metric === 'voltage' || props.metric === 'pf',
      min: props.metric === 'consumption' || props.metric === 'current' ? 0 : undefined,
      splitNumber: props.large ? 5 : 4,
      axisLabel: {
        color: muted,
        fontSize: 12,
        formatter: (v: number) => (props.metric === 'pf' ? formatDecimal(v, 2) : String(v)),
      },
      splitLine: { lineStyle: { color: line } },
    },
    series: [...bandSeries, ...lineSeries],
  }
})
</script>

<template>
  <!-- Inline style: vue-echarts ships an unlayered .echarts height that would beat a Tailwind class. -->
  <VChart
    :option="option"
    autoresize
    :style="{ height: `${height}px`, width: '100%' }"
    role="img"
    :aria-label="description"
  />
</template>
