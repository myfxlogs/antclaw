import { useEffect, useState } from 'react'
import { Search, Ban, Unlock, Shield, Key, Hash } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { listUsers, banUser, unbanUser, adminResetPassword, setUserCodeID } from '../lib/api'

interface User {
  user_id: string
  email: string
  roles: string[]
  banned: boolean
  created_at: string
  code_id: string
}

export default function Users() {
  const { t } = useTranslation()
  const [search, setSearch] = useState('')
  const [users, setUsers] = useState<User[]>([])
  const [loading, setLoading] = useState(true)
  const [showResetModal, setShowResetModal] = useState(false)
  const [selectedUser, setSelectedUser] = useState<User | null>(null)
  const [newPassword, setNewPassword] = useState('')
  const [showCodeIDModal, setShowCodeIDModal] = useState(false)
  const [newCodeID, setNewCodeID] = useState('')
  const [codeIDError, setCodeIDError] = useState('')

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
        code_id: u.code_id || '',
      })))
    } catch (err) {
      console.error('Failed to load users:', err)
    } finally {
      setLoading(false)
    }
  }

  const filteredUsers = users.filter(u => {
    const q = search.toLowerCase()
    return u.email.toLowerCase().includes(q) || (u.code_id || '').includes(q)
  })

  // 前端预校验（后端会再校）：5-10 位、避开 4/7、首位非 0。
  const codeIDRegex = /^[1235689][01235689]{4,9}$/

  const openCodeIDModal = (user: User) => {
    setSelectedUser(user)
    setNewCodeID(user.code_id || '')
    setCodeIDError('')
    setShowCodeIDModal(true)
  }

  const handleSetCodeID = async (autoGenerate: boolean) => {
    if (!selectedUser) return
    const value = autoGenerate ? '' : newCodeID.trim()
    if (value !== '' && !codeIDRegex.test(value)) {
      setCodeIDError('格式错误：5-10 位数字，不含 4/7，首位不能为 0')
      return
    }
    try {
      const r = await setUserCodeID(selectedUser.user_id, value)
      alert(`用户 ${selectedUser.email} 的账号已设为 ${r.code_id}`)
      setShowCodeIDModal(false)
      setSelectedUser(null)
      await loadUsers()
    } catch (err: any) {
      const msg = err?.message || String(err)
      if (msg.includes('already_exists') || msg.includes('already in use')) {
        setCodeIDError('该账号已被其他用户占用')
      } else if (msg.includes('invalid')) {
        setCodeIDError('格式不合法')
      } else {
        setCodeIDError(msg)
      }
    }
  }

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
      alert(t('users.passwordResetSuccess', { email: selectedUser.email }))
      setShowResetModal(false)
      setSelectedUser(null)
      setNewPassword('')
    } catch (err) {
      console.error('Failed to reset password:', err)
      alert(t('users.passwordResetFailed'))
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
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">用户ID</th>
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
                  {user.code_id ? (
                    <span className="font-mono text-sm bg-gray-100 px-2 py-1 rounded">{user.code_id}</span>
                  ) : (
                    <span className="text-xs text-gray-400">未分配</span>
                  )}
                </td>
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
                    <button onClick={() => openCodeIDModal(user)} className="p-2 text-indigo-600 hover:bg-indigo-50 rounded" title="设置/重置用户ID">
                      <Hash className="w-4 h-4" />
                    </button>
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

      {/* CodeID Modal — 设置/重置用户数字账号 */}
      {showCodeIDModal && selectedUser && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl p-6 w-[28rem] shadow-xl">
            <h2 className="text-xl font-bold mb-2">设置用户ID</h2>
            <p className="text-gray-600 mb-3 text-sm">
              用户：<span className="font-mono">{selectedUser.email}</span><br/>
              当前账号：<span className="font-mono">{selectedUser.code_id || '未分配'}</span>
            </p>
            <div className="space-y-3">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  新账号（5-10 位数字，不含 4/7，首位不为 0）
                </label>
                <input
                  type="text"
                  value={newCodeID}
                  onChange={(e) => {
                    setNewCodeID(e.target.value.replace(/[^0-9]/g, ''))
                    setCodeIDError('')
                  }}
                  className="w-full px-4 py-2 border rounded-lg font-mono"
                  placeholder="例如 12356"
                  maxLength={10}
                />
                {codeIDError && <p className="text-red-600 text-xs mt-1">{codeIDError}</p>}
              </div>
              <div className="flex gap-2 justify-between items-center">
                <button
                  onClick={() => handleSetCodeID(true)}
                  className="px-3 py-2 text-sm border rounded-lg hover:bg-gray-50"
                  title="忽略输入，由系统随机分配一个新账号"
                >
                  随机分配
                </button>
                <div className="flex gap-2">
                  <button
                    onClick={() => setShowCodeIDModal(false)}
                    className="px-4 py-2 border rounded-lg hover:bg-gray-50"
                  >
                    取消
                  </button>
                  <button
                    onClick={() => handleSetCodeID(false)}
                    disabled={!newCodeID}
                    className="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 disabled:opacity-50"
                  >
                    确定
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}

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
