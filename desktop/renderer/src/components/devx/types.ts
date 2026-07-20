// Minimal types for the subset of dashforge's dashboardir.Dashboard schema
// this view actually renders (metric, chart-with-line-marks, table). Not a
// full mirror of the dashforge schema — VisionStudio is a read-only
// consumer of whatever devfolio's `devx dashboard` export produces, not an
// implementer of the full dashboard-IR spec.

export interface DashforgeDataSource {
  id: string
  type: string
  data?: unknown
}

export interface DashforgePosition {
  x: number
  y: number
  w: number
  h: number
}

export interface DashforgeMetricConfig {
  valueField: string
  format?: 'number' | 'percent' | 'currency'
  formatOptions?: {
    decimals?: number
    prefix?: string
    suffix?: string
  }
}

export interface DashforgeChartMark {
  id: string
  name?: string
  geometry: string
  encode: { x?: string; y?: string }
  style?: { color?: string }
}

export interface DashforgeChartConfig {
  marks: DashforgeChartMark[]
  legend?: { show: boolean }
}

export interface DashforgeTableColumn {
  field: string
  header?: string
  align?: 'left' | 'right' | 'center'
  format?: string
}

export interface DashforgeTableConfig {
  columns: DashforgeTableColumn[]
}

export interface DashforgeWidget {
  id: string
  title?: string
  type: 'metric' | 'chart' | 'table' | 'text' | 'image'
  position: DashforgePosition
  dataSourceId?: string
  config: DashforgeMetricConfig | DashforgeChartConfig | DashforgeTableConfig | unknown
}

export interface DashforgeDashboard {
  id: string
  title: string
  description?: string
  dataSources: DashforgeDataSource[]
  widgets: DashforgeWidget[]
}
