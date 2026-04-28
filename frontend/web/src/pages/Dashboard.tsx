import { TrendingUp, Calendar, BarChart3, Globe } from 'lucide-react'

// Card component for dashboard sections
function Card({ title, icon: Icon, children, className = '' }: { 
  title: string
  icon: React.ElementType
  children: React.ReactNode
  className?: string 
}) {
  return (
    <div className={`bg-white rounded-xl shadow-sm border p-6 ${className}`}>
      <div className="flex items-center gap-3 mb-4">
        <div className="p-2 bg-blue-50 rounded-lg">
          <Icon className="w-5 h-5 text-blue-600" />
        </div>
        <h2 className="text-lg font-semibold text-gray-900">{title}</h2>
      </div>
      {children}
    </div>
  )
}

// COT Positioning Card
function CotCard() {
  return (
    <Card title="COT Positioning" icon={BarChart3}>
      <div className="space-y-3">
        <div className="flex justify-between items-center">
          <span className="text-gray-600">EUR Net Position</span>
          <span className="font-semibold text-green-600">+12.5K</span>
        </div>
        <div className="flex justify-between items-center">
          <span className="text-gray-600">Weekly Change</span>
          <span className="font-semibold text-green-600">+2.3K</span>
        </div>
        <div className="flex justify-between items-center">
          <span className="text-gray-600">Sentiment</span>
          <span className="px-2 py-1 bg-green-100 text-green-700 rounded-full text-sm">Bullish</span>
        </div>
      </div>
    </Card>
  )
}

// Calendar Events Card
function CalendarCard() {
  const events = [
    { time: '14:00', title: 'Fed Chair Speech', impact: 'high', currency: 'USD' },
    { time: '08:30', title: 'NFP Release', impact: 'high', currency: 'USD' },
    { time: '10:00', title: 'ECB Rate Decision', impact: 'high', currency: 'EUR' },
  ]

  return (
    <Card title="Economic Calendar" icon={Calendar}>
      <div className="space-y-3">
        {events.map((e, i) => (
          <div key={i} className="flex items-center justify-between py-2 border-b last:border-0">
            <div className="flex items-center gap-3">
              <span className="text-sm text-gray-500">{e.time}</span>
              <div>
                <p className="font-medium text-gray-900">{e.title}</p>
                <p className="text-xs text-gray-500">{e.currency}</p>
              </div>
            </div>
            <span className={`px-2 py-1 rounded text-xs font-medium ${
              e.impact === 'high' ? 'bg-red-100 text-red-700' : 'bg-yellow-100 text-yellow-700'
            }`}>
              {e.impact}
            </span>
          </div>
        ))}
      </div>
    </Card>
  )
}

// Macro Regime Card
function MacroCard() {
  return (
    <Card title="Macro Regime" icon={Globe}>
      <div className="space-y-4">
        <div>
          <div className="flex justify-between items-center mb-2">
            <span className="text-gray-600">Current Regime</span>
            <span className="font-semibold text-blue-600">Risk-On</span>
          </div>
          <div className="w-full bg-gray-200 rounded-full h-2">
            <div className="bg-blue-600 h-2 rounded-full" style={{ width: '75%' }}></div>
          </div>
        </div>
        <div className="grid grid-cols-2 gap-3 text-sm">
          <div className="p-3 bg-gray-50 rounded-lg">
            <p className="text-gray-500">US10Y</p>
            <p className="font-semibold">4.25%</p>
          </div>
          <div className="p-3 bg-gray-50 rounded-lg">
            <p className="text-gray-500">DXY</p>
            <p className="font-semibold text-red-600">-0.35%</p>
          </div>
        </div>
      </div>
    </Card>
  )
}

// Signals Overview Card
function SignalsCard() {
  const signals = [
    { pair: 'EURUSD', signal: 'Bullish', strength: 85 },
    { pair: 'GBPUSD', signal: 'Bullish', strength: 78 },
    { pair: 'XAUUSD', signal: 'Neutral', strength: 52 },
  ]

  return (
    <Card title="Signal Overview" icon={TrendingUp} className="col-span-full">
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {signals.map((s) => (
          <div key={s.pair} className="p-4 border rounded-lg">
            <div className="flex justify-between items-center mb-2">
              <span className="font-semibold">{s.pair}</span>
              <span className={`px-2 py-1 rounded text-xs ${
                s.signal === 'Bullish' ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-700'
              }`}>
                {s.signal}
              </span>
            </div>
            <div className="w-full bg-gray-200 rounded-full h-2">
              <div 
                className="bg-blue-600 h-2 rounded-full" 
                style={{ width: `${s.strength}%` }}
              ></div>
            </div>
            <p className="text-right text-xs text-gray-500 mt-1">{s.strength}%</p>
          </div>
        ))}
      </div>
    </Card>
  )
}

export default function Dashboard() {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-900">Dashboard</h1>
        <p className="text-sm text-gray-500">Last updated: {new Date().toLocaleString()}</p>
      </div>
      
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        <CotCard />
        <CalendarCard />
        <MacroCard />
        <SignalsCard />
      </div>
    </div>
  )
}
