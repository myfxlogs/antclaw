import { useEffect, useState } from 'react'
import { Search, Ban, Unlock, Shield, Key } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { listUsers, banUser, unbanUser, adminResetPassword } from '../lib/api'

interface User {
  user_id: string
  email: string
  roles: string[]
  banned: boolean
  created_at: string
}

export default function Users() {
  const { t } = useTranslation()
  const [search, setSearch] = useState('')
  const [users, setUsers] = useState<User[]>([])
  const [loading, setLoading] = useState(true)
  const [showResetModal, setShowResetModal] = useState(false)
  const [selectedUser, setSelectedUser] = useState<User | null>(null)
  const [newPassword, setNewPassword] = useState('')

  useEffect(() => {
    loadUsers()
  }, [])

  const loadUsers = async () => {
    try {
      const response = await listUsers()
      setUsers(response.users.map((u: any) => ({
        user_id: u.user_id,
        email: u.email,
        roles: u.roles && u.roles.length > 0 ? u.roles : ['user'],
        banned: Boolean(u.banned),
        created_at: u.created_at ? new Date(u.created_at * 1000).toISOString().split('T')[0] : '-',
      })))
    } catch (err) {
      console.error('Failed to load users:', err)
    } finally {
      setLoading(false)
    }
  }

  const filteredUsers = users.filter(u => 
    u.email.toLowerCase().includes(search.toLowerCase())
  )

  const handleBan = async (userId: string) => {
    try {
      await banUser(userId, 'Banned by admin', undefined)
      await loadUsers()
    } catch (err) {
      console.error('Failed to ban user:', err)
    }
  }

  const handleUnban = async (userId: string) => {
    try {
      await unbanUser(userId)
      await loadUsers()
    } catch (err) {
      console.error('Failed to unban user:', err)
    }
  }

  const openResetModal = (user: User) => {
    setSelectedUser(user)
    setNewPassword('')
    setShowResetModal(true)
  }

  const handleResetPassword = async () => {
    if (!selectedUser || !newPassword) return
    try {
      await adminResetPassword(selectedUser.user_id, newPassword)
      alert(`Password reset successfully for ${selectedUser.email}`)
      setShowResetModal(false)
      setSelectedUser(null)
      setNewPassword('')
    } catch (err) {
      console.error('Failed to reset password:', err)
      alert('Failed to reset password')
    }
  }

  if (loading) {
    return <div className="flex items-center justify-center h-64">{t('users.loading')}</div>
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-900">{t('users.title')}</h1>
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
          <input
            type="text"
            placeholder={t('users.searchPlaceholder')}
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="pl-10 pr-4 py-2 border rounded-lg w-64"
          />
        </div>
      </div>

      <div className="bg-white rounded-xl shadow-sm overflow-hidden">
        <table className="w-full">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">{t('users.email')}</th>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">{t('users.role')}</th>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">{t('users.status')}</th>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">{t('users.createdAt')}</th>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">{t('users.actions')}</th>
            </tr>
          </thead>
          <tbody className="divide-y">
            {filteredUsers.map((user) => (
              <tr key={user.user_id} className="hover:bg-gray-50">
                <td className="px-6 py-4">
                  <p className="font-medium">{user.email}</p>
                </td>
                <td className="px-6 py-4">
                  <span className={`inline-flex items-center gap-1 px-2 py-1 rounded text-xs font-medium ${
                    user.roles.includes('admin') ? 'bg-purple-100 text-purple-700' : 'bg-gray-100 text-gray-700'
                  }`}>
                    {user.roles.includes('admin') && <Shield className="w-3 h-3" />}
                    {user.roles.join(', ')}
                  </span>
                </td>
                <td className="px-6 py-4">
                  <span className={`inline-flex px-2 py-1 rounded text-xs font-medium ${
                    user.banned ? 'bg-red-100 text-red-700' : 'bg-green-100 text-green-700'
                  }`}>
                    {user.banned ? t('users.banned') : t('users.active')}
                  </span>
                </td>
                <td className="px-6 py-4 text-sm text-gray-500">{user.created_at}</td>
                <td className="px-6 py-4">
                  <div className="flex gap-2">
                    <button onClick={() => openResetModal(user)} className="p-2 text-blue-600 hover:bg-blue-50 rounded" title="Reset Password">
                      <Key className="w-4 h-4" />
                    </button>
                    {!user.banned ? (
                      <button onClick={() => handleBan(user.user_id)} className="p-2 text-red-600 hover:bg-red-50 rounded" title={t('users.ban')}>
                        <Ban className="w-4 h-4" />
                      </button>
                    ) : (
                      <button onClick={() => handleUnban(user.user_id)} className="p-2 text-green-600 hover:bg-green-50 rounded" title={t('users.unban')}>
                        <Unlock className="w-4 h-4" />
                      </button>
                    )}
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Reset Password Modal */}
      {showResetModal && selectedUser && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl p-6 w-96 shadow-xl">
            <h2 className="text-xl font-bold mb-4">Reset Password</h2>
            <p className="text-gray-600 mb-4">User: {selectedUser.email}</p>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">New Password</label>
                <input
                  type="password"
                  value={newPassword}
                  onChange={(e) => setNewPassword(e.target.value)}
                  className="w-full px-4 py-2 border rounded-lg"
                  placeholder="Enter new password"
                />
              </div>
              <div className="flex gap-2 justify-end">
                <button
                  onClick={() => setShowResetModal(false)}
                  className="px-4 py-2 border rounded-lg hover:bg-gray-50"
                >
                  Cancel
                </button>
                <button
                  onClick={handleResetPassword}
                  disabled={!newPassword}
                  className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
                >
                  Reset Password
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
