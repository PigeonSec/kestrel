import { useState, useEffect } from 'react'
import { api } from '../lib/api'
import { Button } from '../components/ui/button'
import { Input } from '../components/ui/input'
import { Label } from '../components/ui/label'
import {
  Card,
  CardHeader,
  CardTitle,
  CardContent,
  CardDescription,
} from '../components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '../components/ui/table'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '../components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '../components/ui/select'
import { Badge } from '../components/ui/badge'
import { Trash2, Plus } from 'lucide-react'
import { useToast } from '../hooks/use-toast'

interface IOC {
  value: string
  type: string
  feed?: string
  stix_id?: string
  misp_event_id?: string
}

export default function IOCs() {
  const [iocs, setIOCs] = useState<IOC[]>([])
  const [loading, setLoading] = useState(true)
  const [dialogOpen, setDialogOpen] = useState(false)
  const { toast } = useToast()

  const [formData, setFormData] = useState({
    ioc_type: 'domain',
    value: '',
    category: 'Malware',
    feed: '',
    comment: '',
    access_level: 'paid',
  })

  useEffect(() => {
    fetchIOCs()
  }, [])

  const fetchIOCs = async () => {
    try {
      const response = await api.get('/api/iocs')
      setIOCs(response.data.iocs || [])
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to fetch IOCs',
        variant: 'destructive',
      })
    } finally {
      setLoading(false)
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    const payload: any = {
      category: formData.category,
      feed: formData.feed,
      comment: formData.comment,
      access_level: formData.access_level,
    }

    // Add the IOC value based on type
    switch (formData.ioc_type) {
      case 'domain':
        payload.domain = formData.value
        break
      case 'ip':
        payload.ip = formData.value
        break
      case 'url':
        payload.url = formData.value
        break
      case 'hash':
        payload.hash = formData.value
        break
      case 'email':
        payload.email = formData.value
        break
    }

    try {
      await api.post('/api/ioc', payload)
      toast({
        title: 'Success',
        description: 'IOC added successfully',
      })
      setDialogOpen(false)
      setFormData({
        ioc_type: 'domain',
        value: '',
        category: 'Malware',
        feed: '',
        comment: '',
        access_level: 'paid',
      })
      fetchIOCs()
    } catch (error: any) {
      toast({
        title: 'Error',
        description: error.response?.data?.error || 'Failed to add IOC',
        variant: 'destructive',
      })
    }
  }

  const handleDelete = async (ioc: IOC) => {
    if (!ioc.feed) {
      toast({
        title: 'Error',
        description: 'Cannot delete IOC: feed information missing',
        variant: 'destructive',
      })
      return
    }

    if (!confirm(`Delete IOC ${ioc.value}?`)) return

    try {
      await api.delete(`/api/ioc/${encodeURIComponent(ioc.value)}?feed=${ioc.feed}`)
      toast({
        title: 'Success',
        description: 'IOC deleted successfully',
      })
      fetchIOCs()
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to delete IOC',
        variant: 'destructive',
      })
    }
  }

  if (loading) {
    return <div className="flex items-center justify-center">Loading...</div>
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold">Indicators of Compromise</h1>
          <p className="text-gray-500 mt-1">Manage your threat intelligence IOCs</p>
        </div>

        <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
          <DialogTrigger asChild>
            <Button>
              <Plus className="h-4 w-4 mr-2" />
              Add IOC
            </Button>
          </DialogTrigger>
          <DialogContent className="max-w-2xl">
            <DialogHeader>
              <DialogTitle>Add New IOC</DialogTitle>
            </DialogHeader>

            <form onSubmit={handleSubmit} className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="ioc_type">IOC Type</Label>
                  <Select
                    value={formData.ioc_type}
                    onValueChange={(value) =>
                      setFormData({ ...formData, ioc_type: value })
                    }
                  >
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="domain">Domain</SelectItem>
                      <SelectItem value="ip">IP Address</SelectItem>
                      <SelectItem value="url">URL</SelectItem>
                      <SelectItem value="hash">Hash</SelectItem>
                      <SelectItem value="email">Email</SelectItem>
                    </SelectContent>
                  </Select>
                </div>

                <div className="space-y-2">
                  <Label htmlFor="value">Value</Label>
                  <Input
                    id="value"
                    value={formData.value}
                    onChange={(e) => setFormData({ ...formData, value: e.target.value })}
                    placeholder={`Enter ${formData.ioc_type}...`}
                    required
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="category">Category</Label>
                  <Select
                    value={formData.category}
                    onValueChange={(value) =>
                      setFormData({ ...formData, category: value })
                    }
                  >
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="Malware">Malware</SelectItem>
                      <SelectItem value="Phishing">Phishing</SelectItem>
                      <SelectItem value="C2">C2</SelectItem>
                      <SelectItem value="Scanning">Scanning</SelectItem>
                    </SelectContent>
                  </Select>
                </div>

                <div className="space-y-2">
                  <Label htmlFor="feed">Feed Name</Label>
                  <Input
                    id="feed"
                    value={formData.feed}
                    onChange={(e) => setFormData({ ...formData, feed: e.target.value })}
                    placeholder="e.g., malware-domains"
                    required
                  />
                </div>
              </div>

              <div className="space-y-2">
                <Label htmlFor="comment">Comment</Label>
                <Input
                  id="comment"
                  value={formData.comment}
                  onChange={(e) => setFormData({ ...formData, comment: e.target.value })}
                  placeholder="Optional description..."
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="access_level">Access Level</Label>
                <Select
                  value={formData.access_level}
                  onValueChange={(value) =>
                    setFormData({ ...formData, access_level: value })
                  }
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="free">Free</SelectItem>
                    <SelectItem value="paid">Paid</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <div className="flex justify-end gap-2">
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => setDialogOpen(false)}
                >
                  Cancel
                </Button>
                <Button type="submit">Add IOC</Button>
              </div>
            </form>
          </DialogContent>
        </Dialog>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>IOC List</CardTitle>
          <CardDescription>
            Total: {iocs.length} indicators
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Value</TableHead>
                <TableHead>Type</TableHead>
                <TableHead>Feed</TableHead>
                <TableHead>STIX ID</TableHead>
                <TableHead>MISP Event</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {iocs.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={6} className="text-center text-gray-500">
                    No IOCs found. Add your first indicator to get started.
                  </TableCell>
                </TableRow>
              ) : (
                iocs.map((ioc, index) => (
                  <TableRow key={index}>
                    <TableCell className="font-mono text-sm">{ioc.value}</TableCell>
                    <TableCell>
                      <Badge variant="outline">{ioc.type}</Badge>
                    </TableCell>
                    <TableCell>
                      <Badge variant="secondary">{ioc.feed || '-'}</Badge>
                    </TableCell>
                    <TableCell className="font-mono text-xs text-gray-500">
                      {ioc.stix_id ? (
                        <span title={ioc.stix_id}>{ioc.stix_id.substring(0, 20)}...</span>
                      ) : (
                        '-'
                      )}
                    </TableCell>
                    <TableCell className="font-mono text-xs text-gray-500">
                      {ioc.misp_event_id ? (
                        <span title={ioc.misp_event_id}>{ioc.misp_event_id.substring(0, 20)}...</span>
                      ) : (
                        '-'
                      )}
                    </TableCell>
                    <TableCell className="text-right">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleDelete(ioc)}
                      >
                        <Trash2 className="h-4 w-4 text-red-500" />
                      </Button>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  )
}
