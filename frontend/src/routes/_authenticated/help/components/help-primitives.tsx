// Reusable primitive components for the Help page layout.
import * as React from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

export function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <Card className="mb-4">
      <CardHeader className="pb-2">
        <CardTitle className="text-base">{title}</CardTitle>
      </CardHeader>
      <CardContent className="text-sm text-muted-foreground space-y-2">{children}</CardContent>
    </Card>
  )
}

export function Steps({ items }: { items: string[] }) {
  return (
    <ol className="list-decimal list-inside space-y-1">
      {items.map((item, i) => <li key={i}>{item}</li>)}
    </ol>
  )
}

export function InfoTable({ rows }: { rows: [string, string][] }) {
  return (
    <table className="w-full text-sm border-collapse">
      <tbody>
        {rows.map(([label, desc]) => (
          <tr key={label} className="border-b last:border-0">
            <td className="py-1.5 pr-4 font-medium text-foreground whitespace-nowrap">{label}</td>
            <td className="py-1.5 text-muted-foreground">{desc}</td>
          </tr>
        ))}
      </tbody>
    </table>
  )
}
