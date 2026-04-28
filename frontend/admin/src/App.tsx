import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import Layout from './components/Layout'
import Dashboard from './pages/Dashboard'
import Users from './pages/Users'
import Jobs from './pages/Jobs'
import Audit from './pages/Audit'
import Settings from './pages/Settings'
import DataSummary from './pages/DataSummary'
import DataSources from './pages/DataSources'
import Strategies from './pages/Strategies'
import SystemAI from './pages/SystemAI'
import Login from './pages/Login'

function PrivateRoute({ children }: { children: React.ReactNode }) {
  const token = localStorage.getItem('token')
  return token ? <>{children}</> : <Navigate to="/login" />
}

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/" element={
          <PrivateRoute>
            <Layout />
          </PrivateRoute>
        }>
          <Route index element={<Dashboard />} />
          <Route path="users" element={<Users />} />
          <Route path="jobs" element={<Jobs />} />
          <Route path="audit" element={<Audit />} />
          <Route path="data" element={<DataSummary />} />
          <Route path="datasources" element={<DataSources />} />
          <Route path="strategies" element={<Strategies />} />
          <Route path="system-ai" element={<SystemAI />} />
          <Route path="settings" element={<Settings />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}

export default App
