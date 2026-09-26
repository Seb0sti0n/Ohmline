<script setup lang="ts">
import { computed } from 'vue'
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import { BarChart } from 'echarts/charts'
import { GridComponent, MarkLineComponent, TooltipComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import { cssColor } from '@/utils/theme'
import { formatDayMonth, formatInt, formatKwh } from '@/utils/format'

use([BarChart, GridComponent, MarkLineComponent, TooltipComponent, CanvasRenderer])

const props = withDefaults(
  defineProps<{
    /** One value per day starting at `from` (YYYY-MM-DD). */
    values: number[]
    from: string
    baseline: number
    /** Days deviating more than this fraction from the baseline are highlighted. */
    threshold?: number
    description: string
  }>(),
  { threshold: 0.2 },
)

const dates = computed(() =>
  props.values.map((_, i) => {
    const d = new Date(`${props.from}T00:00:00Z`)
    d.setUTCDate(d.getUTCDate() + i)
    return d.toISOString().slice(0, 10)
  }),
)

const option = computed(() => {
  const ink = cssColor('ink', '#15201C')
  const muted = cssColor('ink-muted', '#55615C')
  const line = cssColor('line', '#D8DFDB')
  const bar = cssColor('chart-bar', '#A9C7C2')
  const critical = cssColor('status-critical', '#B42318')
  const b = props.baseline

  return {
    animation: false,
    grid: { left: 48, right: 8, top: 10, bottom: 26 },
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'none' },
      formatter: (params: { dataIndex: number }[]) => {
        const i = params[0].dataIndex
        return `${formatDayMonth(dates.value[i])}<br/><b>${formatKwh(props.values[i])}</b><br/>Baseline ${formatKwh(b)}`
      },
    },
    xAxis: {
      type: 'category',
      data: dates.value.map((d) => String(Number(d.slice(8, 10)))),
      axisLine: { show: false },
      axisTick: { show: false },
      axisLabel: { color: muted, fontSize: 12 },
    },
    yAxis: {
      type: 'value',
      min: 0,
      splitNumber: 4,
      axisLabel: { color: muted, fontSize: 12, formatter: (v: number) => formatInt(v) },
      splitLine: { lineStyle: { color: line } },
    },
    series: [
      {
        type: 'bar',
        barWidth: '68%',
        itemStyle: { borderRadius: [3, 3, 0, 0] },
        data: props.values.map((v) => ({
          value: v,
          itemStyle: { color: b > 0 && Math.abs(v / b - 1) > props.threshold ? critical : bar },
        })),
        markLine: {
          silent: true,
          symbol: 'none',
          lineStyle: { color: ink, type: [5, 4], width: 1.5 },
          label: {
            show: true,
            position: 'insideStartTop',
            formatter: `Baseline ${formatKwh(b)}`,
            color: ink,
            fontSize: 12,
          },
          data: [{ yAxis: b }],
        },
      },
    ],
  }
})
</script>

<template>
  <VChart
    :option="option"
    autoresize
    style="height: 200px; width: 100%"
    role="img"
    :aria-label="description"
  />
</template>
