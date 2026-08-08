import ReactECharts from 'echarts-for-react'
import { compile } from '@grokify/echartify'
import type { ChartIR } from '@grokify/echartify'

interface EChartifyProps {
  ir: ChartIR
  height?: number | string
  className?: string
}

export function EChartify({ ir, height = 300, className }: EChartifyProps) {
  const option = compile(ir)

  return (
    <ReactECharts
      option={option}
      style={{ height }}
      className={className}
      opts={{ renderer: 'svg' }}
      theme="dark"
    />
  )
}

export function DonutChart({
  data,
  title,
  height = 200,
}: {
  data: { name: string; value: number }[]
  title?: string
  height?: number
}) {
  const ir: ChartIR = {
    title,
    datasets: [
      {
        id: 'main',
        columns: [
          { name: 'name', type: 'string' },
          { name: 'value', type: 'number' },
        ],
        rows: data.map((d) => [d.name, d.value.toString()]),
      },
    ],
    marks: [
      {
        id: 'pie',
        datasetId: 'main',
        geometry: 'pie',
        encode: { name: 'name', value: 'value' },
      },
    ],
    legend: { show: true, position: 'right' },
    tooltip: { show: true, trigger: 'item' },
  }

  return <EChartify ir={ir} height={height} />
}

export function StackedBarChart({
  data,
  title,
  categoryField,
  series,
  height = 300,
}: {
  data: Record<string, string | number>[]
  title?: string
  categoryField: string
  series: { key: string; name: string }[]
  height?: number
}) {
  const columns = [
    { name: categoryField, type: 'string' as const },
    ...series.map((s) => ({ name: s.key, type: 'number' as const })),
  ]

  const rows = data.map((row) => [
    String(row[categoryField]),
    ...series.map((s) => String(row[s.key] ?? 0)),
  ])

  const ir: ChartIR = {
    title,
    datasets: [
      {
        id: 'main',
        columns,
        rows,
      },
    ],
    marks: series.map((s) => ({
      id: s.key,
      name: s.name,
      datasetId: 'main',
      geometry: 'bar' as const,
      encode: { x: categoryField, y: s.key },
      stack: 'total',
    })),
    axes: [
      { id: 'x', type: 'category', position: 'bottom' },
      { id: 'y', type: 'value', position: 'left' },
    ],
    legend: { show: true, position: 'top' },
    tooltip: { show: true, trigger: 'axis' },
  }

  return <EChartify ir={ir} height={height} />
}
