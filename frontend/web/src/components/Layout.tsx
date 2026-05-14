import { Outlet } from 'react-router-dom'
import Navbar from './Navbar'
import { useSse } from '../features/_shared/useSse'

export default function Layout() {
  useSse()

  return (
    <div className="min-h-screen bg-gray-50">
      <Navbar />
      <main className="container mx-auto px-4 py-6">
        <Outlet />
      </main>
    </div>
  )
}
