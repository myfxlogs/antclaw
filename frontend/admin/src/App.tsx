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
import PushManagement from './pages/PushManagement'
import Devices from './pages/Devices'

// M-A..M-G features
import GEXPage from './features/options/GEXPage'
import IVSurfacePage from './features/options/IVSurfacePage'
import SkewPage from './features/options/SkewPage'
import OptionsAlertsPage from './features/options/AlertsPage'
import MOVEPage from './features/vol/MOVEPage'
import CrossVolPage from './features/vol/CrossVolPage'
import TermStructurePage from './features/vol/TermStructurePage'
import WalkForwardPage from './features/backtest/walkforward/WalkForwardPage'
import FedWatchPage from './features/macro/FedWatchPage'
import MacroExtrasPage from './features/macro/MacroExtrasPage'
import FREDAlertsPage from './features/macro/FREDAlertsPage'
import TreasuryCurvePage from './features/macro/TreasuryCurvePage'
import OnchainPage from './features/onchain/OnchainPage'
import DeFiPage from './features/defi/DeFiPage'
import SECPage from './features/sec/SECPage'
import CBOEPutCallPage from './features/sentiment/CBOEPutCallPage'
import MyFXBookPage from './features/sentiment/MyFXBookPage'
import InsiderTradesPage from './features/sentiment/InsiderTradesPage'
import FinvizPage from './features/sentiment/FinvizPage'
import CryptoSocialPage from './features/sentiment/CryptoSocialPage'
import ChatPage from './features/ai/chat/ChatPage'
import AMTPage from './features/ta/amt/AMTPage'
import VolumeProfilePage from './features/microstructure/vp/VolumeProfilePage'
import RegimeOverlayPage from './features/signals/regime/RegimeOverlayPage'

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
          <Route path="push" element={<PushManagement />} />
          <Route path="devices" element={<Devices />} />

          {/* features */}
          <Route path="options/gex" element={<GEXPage />} />
          <Route path="options/iv-surface" element={<IVSurfacePage />} />
          <Route path="options/skew" element={<SkewPage />} />
          <Route path="options/alerts" element={<OptionsAlertsPage />} />
          <Route path="vol/move" element={<MOVEPage />} />
          <Route path="vol/cross" element={<CrossVolPage />} />
          <Route path="vol/term" element={<TermStructurePage />} />
          <Route path="backtest/walkforward" element={<WalkForwardPage />} />
          <Route path="macro/fedwatch" element={<FedWatchPage />} />
          <Route path="macro/extras" element={<MacroExtrasPage />} />
          <Route path="macro/fred-alerts" element={<FREDAlertsPage />} />
          <Route path="macro/treasury" element={<TreasuryCurvePage />} />
          <Route path="onchain" element={<OnchainPage />} />
          <Route path="defi" element={<DeFiPage />} />
          <Route path="sec" element={<SECPage />} />
          <Route path="sentiment/cboe-pc" element={<CBOEPutCallPage />} />
          <Route path="sentiment/myfxbook" element={<MyFXBookPage />} />
          <Route path="sentiment/insider" element={<InsiderTradesPage />} />
          <Route path="sentiment/finviz" element={<FinvizPage />} />
          <Route path="sentiment/crypto-social" element={<CryptoSocialPage />} />
          <Route path="ai/chat" element={<ChatPage />} />
          <Route path="ta/amt" element={<AMTPage />} />
          <Route path="microstructure/vp" element={<VolumeProfilePage />} />
          <Route path="signals/regime" element={<RegimeOverlayPage />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}

export default App
