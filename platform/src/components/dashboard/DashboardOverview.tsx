import { useEffect, useMemo, useState } from 'react'
import { createBrowserClient } from '@/lib/supabase'
import { formatTokenCount, bytesToTokens } from '@/lib/types'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { IconTrendingUp, IconTerminal2, IconBolt, IconCoins } from '@tabler/icons-react'

interface Stats {
  savedBytes: number
  requestCount: number
  avgReduction: number
  totalCredits: number
}

interface ToolCall {
  id: string
  tool_type: string
  integration: string
  optimizer: string
  raw_bytes: number
  optimized_bytes: number
  reduction_ratio: number
  was_cached: boolean
  created_at: string
}

export function DashboardOverview({ userId }: { userId: string }) {
  // useMemo: client created synchronously, stable across renders, no first-render null.
  const supabase = useMemo(() => createBrowserClient(), [])

  const [stats,   setStats]   = useState<Stats | null>(null)
  const [rows,    setRows]    = useState<ToolCall[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (!userId) return
    const month = new Date().toISOString().slice(0, 7)

    Promise.all([
      supabase.from('monthly_usage').select('*').eq('user_id', userId).eq('month', month).maybeSingle(),
      supabase.from('credit_balances').select('total_credits').eq('user_id', userId).maybeSingle(),
      supabase.from('tool_call_summaries').select('*').eq('user_id', userId).order('created_at', { ascending: false }).limit(20),
    ]).then(([{ data: usage }, { data: credits }, { data: recent }]) => {
      const recentRows = recent ?? []
      setStats({
        savedBytes:    usage?.saved_bytes ?? 0,
        requestCount:  usage?.request_count ?? 0,
        avgReduction:  recentRows.length > 0
          ? recentRows.reduce((s, r) => s + (r.reduction_ratio ?? 0), 0) / recentRows.length
          : 0,
        totalCredits: credits?.total_credits ?? 0,
      })
      setRows(recentRows)
      setLoading(false)
    })
  }, [userId, supabase])

  if (loading) return <OverviewSkeleton />

  const savedTokens = bytesToTokens(stats?.savedBytes ?? 0)
  const hasData     = (stats?.requestCount ?? 0) > 0

  if (!hasData) return <EmptyDashboard />

  return (
    <>
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard title="Context saved" value={formatTokenCount(savedTokens) + ' tokens'} sub="this month" icon={<IconBolt className="size-4 text-muted-foreground" />} />
        <StatCard title="Optimized calls" value={String(stats?.requestCount ?? 0)} sub="this month" icon={<IconTrendingUp className="size-4 text-muted-foreground" />} />
        <StatCard title="Avg reduction" value={`${Math.round((stats?.avgReduction ?? 0) * 100)}%`} sub="across all tools" icon={<IconTerminal2 className="size-4 text-muted-foreground" />} />
        <StatCard title="Credits" value={String(stats?.totalCredits ?? 0)} sub="remaining" icon={<IconCoins className="size-4 text-muted-foreground" />} />
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Recent activity</CardTitle>
          <CardDescription>Latest optimized tool calls</CardDescription>
        </CardHeader>
        <CardContent>
          <ActivityTable rows={rows} />
        </CardContent>
      </Card>
    </>
  )
}

function StatCard({ title, value, sub, icon }: { title: string; value: string; sub: string; icon: React.ReactNode }) {
  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium">{title}</CardTitle>
        {icon}
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold">{value}</div>
        <p className="text-xs text-muted-foreground">{sub}</p>
      </CardContent>
    </Card>
  )
}

function ActivityTable({ rows }: { rows: ToolCall[] }) {
  if (rows.length === 0) return <p className="text-sm text-muted-foreground py-4 text-center">No tool calls yet.</p>
  return (
    <div className="overflow-x-auto">
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b text-left text-muted-foreground">
            <th className="pb-2 pr-4">Tool</th>
            <th className="pb-2 pr-4">Raw</th>
            <th className="pb-2 pr-4">Optimized</th>
            <th className="pb-2 pr-4">Saved</th>
            <th className="pb-2 pr-4">Mode</th>
            <th className="pb-2">Time</th>
          </tr>
        </thead>
        <tbody>
          {rows.map(r => {
            const pct      = Math.round((r.reduction_ratio ?? 0) * 100)
            const isRemote = !['bash','read','webfetch','generic'].includes(r.tool_type ?? '')
            return (
              <tr key={r.id} className="border-b last:border-0">
                <td className="py-2 pr-4 font-medium capitalize">{r.tool_type}</td>
                <td className="py-2 pr-4 text-muted-foreground">{formatTokenCount(bytesToTokens(r.raw_bytes))}k</td>
                <td className="py-2 pr-4 text-muted-foreground">{formatTokenCount(bytesToTokens(r.optimized_bytes))}k</td>
                <td className="py-2 pr-4"><span className="text-green-500 font-medium">{pct}%</span></td>
                <td className="py-2 pr-4">
                  <Badge variant={isRemote ? 'default' : 'secondary'} className="text-xs">
                    {isRemote ? 'remote' : 'local'}
                  </Badge>
                </td>
                <td className="py-2 text-muted-foreground text-xs">
                  {new Date(r.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                </td>
              </tr>
            )
          })}
        </tbody>
      </table>
    </div>
  )
}

function OverviewSkeleton() {
  return (
    <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
      {[...Array(4)].map((_, i) => (
        <Card key={i}><CardHeader className="pb-2"><Skeleton className="h-4 w-24" /></CardHeader><CardContent><Skeleton className="h-8 w-16" /></CardContent></Card>
      ))}
    </div>
  )
}

function EmptyDashboard() {
  return (
    <Card className="flex flex-col items-center justify-center py-16 text-center">
      <CardHeader>
        <CardTitle>No data yet</CardTitle>
        <CardDescription>Install Confire and run Claude Code to see your first optimization.</CardDescription>
      </CardHeader>
      <CardContent className="space-y-2">
        <pre className="rounded bg-muted px-4 py-2 text-sm inline-block"><code>brew install confire-dev/tap/confire</code></pre>
        <br />
        <pre className="rounded bg-muted px-4 py-2 text-sm inline-block"><code>confire setup && confire login</code></pre>
      </CardContent>
    </Card>
  )
}
