<script setup lang="ts">
import { computed } from 'vue'
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import { BarChart } from 'echarts/charts'
import { GridComponent, MarkLineComponent, TooltipComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import type { DailyPoint } from '@/types/api'
import { cssColor } from '@/utils/theme'
import { formatDay, formatDayMonth, formatInt, formatKwh } from '@/utils/format'

use([BarChart, GridComponent, MarkLineComponent, TooltipComponent, CanvasRenderer])

const props = defineProps<{ points: DailyPoint[] }>()

/** A day is highlighted when it is more than this above the baseline. */
const THRESHOLD = 0.03

const baseline = computed(() => props.points[0]?.baseline_kwh ?? 0)

const option = computed(() => {
  const ink = cssColor('ink', '#15201C')
  const muted = cssColor('ink-muted', '#55615C')
  const line = cssColor('line', '#D8DFDB')
  const bar = cssColor('chart-bar', '#A9C7C2')
  const brand = cssColor('brand', '#0E6B63')

  const values = props.points.map((p) => p.consumption_kwh)
  const b = baseline.value
  // Bars start below the lowest value so the differences are readable (as in the approved design).
  const min = Math.floor((Math.min(...values, b) - 1000) / 1000) * 1000
  const max = Math.ceil(Math.max(...values, b) / 1000) * 1000

  return {
    animation: false,
    grid: { left: 44, right: 8, top: 10, bottom: 26 },
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'none' },
      formatter: (params: { dataIndex: number }[]) => {
        const p = props.points[params[0].dataIndex]
        return `${formatDayMonth(p.date)}<br/><b>${formatKwh(p.consumption_kwh)}</b><br/>Baseline ${formatKwh(p.baseline_kwh)}`
      },
    },
    xAxis: {
      type: 'category',
      data: props.points.map((p) => String(formatDay(p.date))),
      axisLine: { show: false },
      axisTick: { show: false },
      axisLabel: { color: muted, fontSize: 12 },
    },
    yAxis: {
      type: 'value',
      min,
      max,
      interval: 1000,
      axisLabel: {
        color: muted,
        fontSize: 12,
        formatter: (v: number) => `${formatInt(v / 1000)}k`,
      },
      splitLine: { lineStyle: { color: line } },
    },
    series: [
      {
        type: 'bar',
        barWidth: '72%',
        itemStyle: { borderRadius: [3, 3, 0, 0] },
        data: values.map((v) => ({
          value: v,
          itemStyle: { color: v > b * (1 + THRESHOLD) ? brand : bar },
        })),
        markLine: {
          silent: true,
          symbol: 'none',
          label: { show: false },
          lineStyle: { color: ink, type: [5, 4], width: 1.5 },
          data: [{ yAxis: b }],
        },
      },
    ],
  }
})

const summary = computed(() => {
  const above = props.points.filter(
    (p) => p.consumption_kwh > p.baseline_kwh * (1 + THRESHOLD),
  ).length
  return `Consumo total diario de ${props.points.length} días frente al baseline de ${formatKwh(baseline.value)} por día: ${above} días por encima.`
})
</script>

<template>
  <!-- Inline style: vue-echarts ships an unlayered .echarts height that would beat a Tailwind class. -->
  <VChart
    :option="option"
    autoresize
    style="height: 230px; width: 100%"
    role="img"
    :aria-label="summary"
  />
</template>
