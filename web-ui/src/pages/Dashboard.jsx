import { useEffect, useState } from 'react'
import { useAuth } from '../context/AuthContext'
import { bankingAPI } from '../services/api'
import { Wallet, ArrowUpRight, ArrowDownLeft, TrendingUp, Activity } from 'lucide-react'

export default function Dashboard() {
  const { user } = useAuth()
  const [balance, setBalance] = useState(null)
  const [recentTransactions, setRecentTransactions] = useState([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    loadDashboardData()
  }, [])

  const loadDashboardData = async () => {
    try {
      setLoading(true)
      // Load balance
      const balanceData = await bankingAPI.getBalance(user.id, 'ACC_001', user.channel)
      setBalance(balanceData)

      // Load recent transactions
      const historyData = await bankingAPI.getTransactionHistory(user.id, 30)
      setRecentTransactions(historyData.transactions?.slice(0, 5) || [])
    } catch (error) {
      console.error('Failed to load dashboard data:', error)
    } finally {
      setLoading(false)
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-gray-800">Welcome back, {user.name}</h1>
        <p className="text-gray-600 mt-2">Here's your banking overview</p>
      </div>

      {/* Balance Card */}
      <div className="bg-gradient-to-r from-primary-600 to-primary-800 rounded-2xl shadow-lg p-8 text-white">
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center space-x-3">
            <Wallet size={32} />
            <div>
              <p className="text-primary-200 text-sm">Total Balance</p>
              <p className="text-3xl font-bold">
                ₹{balance?.balance?.toLocaleString('en-IN') || '0.00'}
              </p>
            </div>
          </div>
        </div>
        <div className="mt-4 pt-4 border-t border-primary-500">
          <p className="text-sm text-primary-200">Account: {balance?.account_id || 'ACC_001'}</p>
        </div>
      </div>

      {/* Quick Stats */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div className="bg-white rounded-xl shadow-md p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-600 text-sm">Total Income</p>
              <p className="text-2xl font-bold text-green-600 mt-1">₹45,000</p>
            </div>
            <div className="bg-green-100 p-3 rounded-lg">
              <ArrowDownLeft className="text-green-600" size={24} />
            </div>
          </div>
        </div>

        <div className="bg-white rounded-xl shadow-md p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-600 text-sm">Total Expenses</p>
              <p className="text-2xl font-bold text-red-600 mt-1">₹12,500</p>
            </div>
            <div className="bg-red-100 p-3 rounded-lg">
              <ArrowUpRight className="text-red-600" size={24} />
            </div>
          </div>
        </div>

        <div className="bg-white rounded-xl shadow-md p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-600 text-sm">Savings Rate</p>
              <p className="text-2xl font-bold text-primary-600 mt-1">72%</p>
            </div>
            <div className="bg-primary-100 p-3 rounded-lg">
              <TrendingUp className="text-primary-600" size={24} />
            </div>
          </div>
        </div>
      </div>

      {/* Recent Transactions */}
      <div className="bg-white rounded-xl shadow-md p-6">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-xl font-bold text-gray-800 flex items-center space-x-2">
            <Activity size={24} />
            <span>Recent Transactions</span>
          </h2>
        </div>
        <div className="space-y-3">
          {recentTransactions.length > 0 ? (
            recentTransactions.map((txn, index) => (
              <div
                key={index}
                className="flex items-center justify-between p-4 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors"
              >
                <div className="flex items-center space-x-4">
                  <div
                    className={`p-2 rounded-lg ${
                      txn.type === 'credit'
                        ? 'bg-green-100 text-green-600'
                        : 'bg-red-100 text-red-600'
                    }`}
                  >
                    {txn.type === 'credit' ? (
                      <ArrowDownLeft size={20} />
                    ) : (
                      <ArrowUpRight size={20} />
                    )}
                  </div>
                  <div>
                    <p className="font-semibold text-gray-800">{txn.description || 'Transaction'}</p>
                    <p className="text-sm text-gray-500">{txn.date || 'N/A'}</p>
                  </div>
                </div>
                <div className="text-right">
                  <p
                    className={`font-bold ${
                      txn.type === 'credit' ? 'text-green-600' : 'text-red-600'
                    }`}
                  >
                    {txn.type === 'credit' ? '+' : '-'}₹{txn.amount?.toLocaleString('en-IN') || '0'}
                  </p>
                  <p className="text-sm text-gray-500">{txn.status || 'Completed'}</p>
                </div>
              </div>
            ))
          ) : (
            <p className="text-center text-gray-500 py-8">No recent transactions</p>
          )}
        </div>
      </div>
    </div>
  )
}

