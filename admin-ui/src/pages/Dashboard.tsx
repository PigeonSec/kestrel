import { useState, useEffect } from 'react'
import { api } from '../lib/api'
import { Card, CardHeader, CardTitle, CardContent } from '../components/ui/card'
import { AlertTriangle, Rss, Key, Shield } from 'lucide-react'

interface Stats {
  totalIOCs: number
  totalFeeds: number
  totalKeys: number
}

export default function Dashboard() {
  const [stats, setStats] = useState<Stats>({
    totalIOCs: 0,
    totalFeeds: 0,
    totalKeys: 0,
  })

  useEffect(() => {
    fetchStats()
  }, [])

  const fetchStats = async () => {
    try {
      const [iocsRes, feedsRes] = await Promise.all([
        api.get('/api/iocs'),
        api.get('/api/feeds'),
      ])

      setStats({
        totalIOCs: iocsRes.data.count || 0,
        totalFeeds: feedsRes.data.feeds?.length || 0,
        totalKeys: 0, // TODO: Implement API keys count
      })
    } catch (error) {
      console.error('Failed to fetch stats:', error)
    }
  }

  const statCards = [
    {
      title: 'Total IOCs',
      value: stats.totalIOCs,
      icon: AlertTriangle,
      color: 'text-red-500',
    },
    {
      title: 'Active Feeds',
      value: stats.totalFeeds,
      icon: Rss,
      color: 'text-blue-500',
    },
    {
      title: 'API Keys',
      value: stats.totalKeys,
      icon: Key,
      color: 'text-green-500',
    },
  ]

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">Dashboard</h1>
        <p className="text-gray-500 mt-1">Overview of your Kestrel CTI platform</p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        {statCards.map((stat) => (
          <Card key={stat.title}>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">{stat.title}</CardTitle>
              <stat.icon className={`h-4 w-4 ${stat.color}`} />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stat.value}</div>
            </CardContent>
          </Card>
        ))}
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Shield className="h-5 w-5" />
            Kestrel CTI Platform
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-2">
          <p className="text-sm text-gray-600">
            Your threat intelligence platform is operational. Use the navigation above to:
          </p>
          <ul className="list-disc list-inside text-sm text-gray-600 space-y-1">
            <li>Manage Indicators of Compromise (IOCs)</li>
            <li>Configure and monitor threat feeds</li>
            <li>Generate and manage API keys for integrations</li>
          </ul>
          <div className="mt-4 pt-4 border-t">
            <p className="text-xs text-gray-500">
              Fully compliant with STIX 2.1, TAXII 2.1, and MISP standards
            </p>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
