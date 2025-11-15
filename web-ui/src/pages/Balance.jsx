import { useEffect, useState } from 'react'
import { useAuth } from '../context/AuthContext'
import { bankingAPI } from '../services/api'
import { Wallet, RefreshCw } from 'lucide-react'

export default function Balance() {
  const { user } = useAuth()
  const [balance, setBalance] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  useEffect(() => {
    loadBalance()
  }, [])

  const loadBalance = async () => {
    try {
      setLoading(true)
      setError(null)
      const data = await bankingAPI.getBalance(user.id, 'ACC_001', user.channel)
      setBalance(data)
    } catch (err) {
      setError('Failed to load balance. Please try again.')
      console.error('Balance error:', err)
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
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-800">Account Balance</h1>
          <p className="text-gray-600 mt-2">View your account balance and details</p>
        </div>
        <button
          onClick={loadBalance}
          className="flex items-center space-x-2 px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition-colors"
        >
          <RefreshCw size={20} />
          <span>Refresh</span>
        </button>
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg">
          {error}
        </div>
      )}

      {balance && (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div className="bg-gradient-to-r from-primary-600 to-primary-800 rounded-2xl shadow-lg p-8 text-white">
            <div className="flex items-center space-x-3 mb-6">
              <Wallet size={32} />
              <div>
                <p className="text-primary-200 text-sm">Available Balance</p>
                <p className="text-4xl font-bold mt-1">
                  ₹{balance.balance?.toLocaleString('en-IN') || '0.00'}
                </p>
              </div>
            </div>
            <div className="space-y-2 pt-4 border-t border-primary-500">
              <div className="flex justify-between">
                <span className="text-primary-200">Account Number</span>
                <span className="font-semibold">{balance.account_id || 'ACC_001'}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-primary-200">Account Type</span>
                <span className="font-semibold">{balance.account_type || 'Savings'}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-primary-200">Currency</span>
                <span className="font-semibold">{balance.currency || 'INR'}</span>
              </div>
            </div>
          </div>

          <div className="bg-white rounded-xl shadow-md p-6">
            <h3 className="text-lg font-semibold text-gray-800 mb-4">Account Details</h3>
            <div className="space-y-3">
              <div className="flex justify-between py-2 border-b">
                <span className="text-gray-600">Account Holder</span>
                <span className="font-semibold">{user.name}</span>
              </div>
              <div className="flex justify-between py-2 border-b">
                <span className="text-gray-600">Channel</span>
                <span className="font-semibold">{balance.channel || user.channel}</span>
              </div>
              <div className="flex justify-between py-2 border-b">
                <span className="text-gray-600">Last Updated</span>
                <span className="font-semibold">
                  {balance.last_updated || new Date().toLocaleDateString()}
                </span>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

