import { useState, useEffect } from 'react'
import { api } from '../lib/api'
import { Card, CardHeader, CardTitle, CardContent } from '../components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '../components/ui/table'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '../components/ui/select'
import { Badge } from '../components/ui/badge'
import { useToast } from '../hooks/use-toast'

interface Feed {
  name: string
  count: number
  access_level?: string
}

export default function Feeds() {
  const [feeds, setFeeds] = useState<Feed[]>([])
  const [loading, setLoading] = useState(true)
  const { toast } = useToast()

  useEffect(() => {
    fetchFeeds()
  }, [])

  const fetchFeeds = async () => {
    try {
      const response = await api.get('/api/feeds')
      setFeeds(response.data.feeds || [])
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to fetch feeds',
        variant: 'destructive',
      })
    } finally {
      setLoading(false)
    }
  }

  const handleAccessLevelChange = async (feedName: string, newLevel: string) => {
    try {
      await api.put(`/api/feeds/${encodeURIComponent(feedName)}/permissions`, {
        access_level: newLevel,
      })
      toast({
        title: 'Success',
        description: 'Feed permissions updated',
      })
      fetchFeeds()
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to update feed permissions',
        variant: 'destructive',
      })
    }
  }

  if (loading) {
    return <div className="flex items-center justify-center">Loading...</div>
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">Threat Intelligence Feeds</h1>
        <p className="text-gray-500 mt-1">Manage your threat intelligence feeds</p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Active Feeds</CardTitle>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Feed Name</TableHead>
                <TableHead>IOC Count</TableHead>
                <TableHead>Access Level</TableHead>
                <TableHead>Endpoints</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {feeds.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={4} className="text-center text-gray-500">
                    No feeds found. Add IOCs to create feeds automatically.
                  </TableCell>
                </TableRow>
              ) : (
                feeds.map((feed) => (
                  <TableRow key={feed.name}>
                    <TableCell className="font-medium">{feed.name}</TableCell>
                    <TableCell>{feed.count}</TableCell>
                    <TableCell>
                      <Select
                        value={feed.access_level || 'paid'}
                        onValueChange={(value) => handleAccessLevelChange(feed.name, value)}
                      >
                        <SelectTrigger className="w-[130px]">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="free">
                            <span className="flex items-center gap-2">
                              <Badge variant="secondary">Free</Badge>
                            </span>
                          </SelectItem>
                          <SelectItem value="paid">
                            <span className="flex items-center gap-2">
                              <Badge variant="default">Paid</Badge>
                            </span>
                          </SelectItem>
                          <SelectItem value="private">
                            <span className="flex items-center gap-2">
                              <Badge variant="destructive">Private</Badge>
                            </span>
                          </SelectItem>
                        </SelectContent>
                      </Select>
                    </TableCell>
                    <TableCell className="text-xs font-mono text-gray-500">
                      /feeds/{feed.name}
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Feed Integration</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div>
            <h3 className="text-sm font-medium mb-2">TAXII 2.1 Endpoint</h3>
            <code className="text-xs bg-gray-100 p-2 rounded block">
              GET /taxii2/api1/collections/
            </code>
          </div>
          <div>
            <h3 className="text-sm font-medium mb-2">MISP Format</h3>
            <code className="text-xs bg-gray-100 p-2 rounded block">
              GET /misp/events
            </code>
          </div>
          <div>
            <h3 className="text-sm font-medium mb-2">STIX 2.1 Bundle</h3>
            <code className="text-xs bg-gray-100 p-2 rounded block">
              GET /stix/bundle
            </code>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
