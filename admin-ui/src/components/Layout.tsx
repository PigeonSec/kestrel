import { Outlet, Link, useNavigate } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'
import { Button } from './ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from './ui/dropdown-menu'
import { Shield, AlertTriangle, Rss, Key, LayoutDashboard, LogOut, Menu } from 'lucide-react'

export default function Layout() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <nav className="bg-white shadow-sm border-b">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16">
            <div className="flex">
              <div className="flex-shrink-0 flex items-center">
                <Shield className="h-8 w-8 text-primary" />
                <span className="ml-2 text-xl font-bold">Kestrel Admin</span>
              </div>
              <div className="hidden sm:ml-6 sm:flex sm:space-x-4">
                <Link
                  to="/"
                  className="inline-flex items-center px-3 py-2 text-sm font-medium rounded-md hover:bg-gray-100"
                >
                  <LayoutDashboard className="h-4 w-4 mr-2" />
                  Dashboard
                </Link>
                <Link
                  to="/iocs"
                  className="inline-flex items-center px-3 py-2 text-sm font-medium rounded-md hover:bg-gray-100"
                >
                  <AlertTriangle className="h-4 w-4 mr-2" />
                  IOCs
                </Link>
                <Link
                  to="/feeds"
                  className="inline-flex items-center px-3 py-2 text-sm font-medium rounded-md hover:bg-gray-100"
                >
                  <Rss className="h-4 w-4 mr-2" />
                  Feeds
                </Link>
                <Link
                  to="/api-keys"
                  className="inline-flex items-center px-3 py-2 text-sm font-medium rounded-md hover:bg-gray-100"
                >
                  <Key className="h-4 w-4 mr-2" />
                  API Keys
                </Link>
              </div>
            </div>

            <div className="flex items-center">
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button variant="ghost" size="sm">
                    <Menu className="h-5 w-5 mr-2" />
                    {user?.username}
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  <DropdownMenuItem onClick={handleLogout}>
                    <LogOut className="h-4 w-4 mr-2" />
                    Logout
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </div>
          </div>
        </div>
      </nav>

      <main className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
        <Outlet />
      </main>
    </div>
  )
}
