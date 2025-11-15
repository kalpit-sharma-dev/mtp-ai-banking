import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom'
import Layout from './components/Layout'
import Dashboard from './pages/Dashboard'
import Transfer from './pages/Transfer'
import Balance from './pages/Balance'
import Statement from './pages/Statement'
import Beneficiaries from './pages/Beneficiaries'
import AIAssistant from './pages/AIAssistant'
import { AuthProvider } from './context/AuthContext'

function App() {
  return (
    <AuthProvider>
      <Router>
        <Layout>
          <Routes>
            <Route path="/" element={<Navigate to="/dashboard" replace />} />
            <Route path="/dashboard" element={<Dashboard />} />
            <Route path="/balance" element={<Balance />} />
            <Route path="/transfer" element={<Transfer />} />
            <Route path="/statement" element={<Statement />} />
            <Route path="/beneficiaries" element={<Beneficiaries />} />
            <Route path="/ai-assistant" element={<AIAssistant />} />
          </Routes>
        </Layout>
      </Router>
    </AuthProvider>
  )
}

export default App

