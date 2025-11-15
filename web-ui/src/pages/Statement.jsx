import { useEffect, useState } from 'react'
import { useAuth } from '../context/AuthContext'
import { bankingAPI } from '../services/api'
import { FileText, Download, Calendar } from 'lucide-react'

export default function Statement() {
  const { user } = useAuth()
  const [transactions, setTransactions] = useState([])
  const [loading, setLoading] = useState(true)
  const [dateRange, setDateRange] = useState({
    from: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
    to: new Date().toISOString().split('T')[0],
  })

  useEffect(() => {
    loadStatement()
  }, [dateRange])

  const loadStatement = async () => {
    try {
      setLoading(true)
      // Use statement API with date range
      const fromDate = new Date(dateRange.from).toISOString()
      const toDate = new Date(dateRange.to).toISOString()
      const data = await bankingAPI.getStatement(user.id, 'ACC_001', fromDate, toDate, user.channel)
      setTransactions(data.transactions || [])
    } catch (error) {
      console.error('Failed to load statement:', error)
      // Fallback to transaction history if statement fails
      try {
        const days = Math.ceil((new Date(dateRange.to) - new Date(dateRange.from)) / (1000 * 60 * 60 * 24))
        const historyData = await bankingAPI.getTransactionHistory(user.id, days)
        setTransactions(historyData.transactions || [])
      } catch (fallbackError) {
        console.error('Failed to load transaction history:', fallbackError)
        setTransactions([])
      }
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
          <h1 className="text-3xl font-bold text-gray-800">Account Statement</h1>
          <p className="text-gray-600 mt-2">View your transaction history</p>
        </div>
        <button className="flex items-center space-x-2 px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition-colors">
          <Download size={20} />
          <span>Download PDF</span>
        </button>
      </div>

      <div className="bg-white rounded-xl shadow-md p-6">
        <div className="flex items-center space-x-4 mb-6">
          <Calendar size={20} className="text-gray-600" />
          <div className="flex space-x-2">
            <input
              type="date"
              value={dateRange.from}
              onChange={(e) => setDateRange({ ...dateRange, from: e.target.value })}
              className="px-3 py-2 border border-gray-300 rounded-lg"
            />
            <span className="self-center text-gray-500">to</span>
            <input
              type="date"
              value={dateRange.to}
              onChange={(e) => setDateRange({ ...dateRange, to: e.target.value })}
              className="px-3 py-2 border border-gray-300 rounded-lg"
            />
          </div>
        </div>

        <div className="space-y-2">
          {transactions.length > 0 ? (
            transactions.map((txn, index) => (
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
                    <FileText size={20} />
                  </div>
                  <div>
                    <p className="font-semibold text-gray-800">{txn.description || 'Transaction'}</p>
                    <p className="text-sm text-gray-500">{txn.date || 'N/A'}</p>
                    {txn.reference && (
                      <p className="text-xs text-gray-400">Ref: {txn.reference}</p>
                    )}
                  </div>
                </div>
                <div className="text-right">
                  <p
                    className={`font-bold text-lg ${
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
            <p className="text-center text-gray-500 py-8">No transactions found</p>
          )}
        </div>
      </div>
    </div>
  )
}

