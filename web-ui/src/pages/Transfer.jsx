import { useState } from 'react'
import { useAuth } from '../context/AuthContext'
import { orchestratorAPI } from '../services/api'
import { ArrowRightLeft, CheckCircle, XCircle } from 'lucide-react'

export default function Transfer() {
  const { user } = useAuth()
  const [formData, setFormData] = useState({
    to_account: '',
    ifsc: '',
    amount: '',
    transfer_type: 'NEFT',
    remarks: '',
  })
  const [loading, setLoading] = useState(false)
  const [result, setResult] = useState(null)
  const [error, setError] = useState(null)

  const handleSubmit = async (e) => {
    e.preventDefault()
    setLoading(true)
    setError(null)
    setResult(null)

    try {
      // Use AI Orchestrator for natural language processing
      const transferText = `Transfer ${formData.amount} rupees to account ${formData.to_account} via ${formData.transfer_type}${formData.ifsc ? ` with IFSC ${formData.ifsc}` : ''}${formData.remarks ? `. Remarks: ${formData.remarks}` : ''}`
      
      const response = await orchestratorAPI.processRequest(transferText, user.id, user.channel)

      if (response.status === 'APPROVED' || response.final_result?.status === 'APPROVED') {
        const finalResult = response.final_result || {}
        const transactionId = finalResult.transaction_id || response.transaction_id
        const amount = finalResult.amount || parseFloat(formData.amount)
        
        setResult({
          success: true,
          transaction_id: transactionId,
          message: response.explanation || finalResult.message || 'Transfer completed successfully',
          risk_score: response.risk_score,
        })
        setFormData({
          to_account: '',
          ifsc: '',
          amount: '',
          transfer_type: 'NEFT',
          remarks: '',
        })
        
        // Trigger refresh events for other pages
        window.dispatchEvent(new CustomEvent('transactionCompleted', { 
          detail: { 
            transactionId,
            amount,
            toAccount: finalResult.to_account || formData.to_account 
          } 
        }))
        window.dispatchEvent(new CustomEvent('balanceUpdated'))
      } else {
        setError(response.explanation || response.final_result?.message || 'Transfer was not approved')
      }
    } catch (err) {
      setError(err.message || err.response?.data?.error || 'Failed to process transfer. Please try again.')
      console.error('Transfer error:', err)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-gray-800">Fund Transfer</h1>
        <p className="text-gray-600 mt-2">Transfer money to any bank account</p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2">
          <div className="bg-white rounded-xl shadow-md p-6">
            <form onSubmit={handleSubmit} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Transfer Type
                </label>
                <select
                  value={formData.transfer_type}
                  onChange={(e) => setFormData({ ...formData, transfer_type: e.target.value })}
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                  required
                >
                  <option value="NEFT">NEFT</option>
                  <option value="RTGS">RTGS</option>
                  <option value="IMPS">IMPS</option>
                  <option value="UPI">UPI</option>
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Beneficiary Account Number
                </label>
                <input
                  type="text"
                  value={formData.to_account}
                  onChange={(e) => setFormData({ ...formData, to_account: e.target.value })}
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                  placeholder="Enter account number"
                  required
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">IFSC Code</label>
                <input
                  type="text"
                  value={formData.ifsc}
                  onChange={(e) => setFormData({ ...formData, ifsc: e.target.value.toUpperCase() })}
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                  placeholder="BANK0001234"
                  required
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">Amount (₹)</label>
                <input
                  type="number"
                  value={formData.amount}
                  onChange={(e) => setFormData({ ...formData, amount: e.target.value })}
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                  placeholder="0.00"
                  min="1"
                  step="0.01"
                  required
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">Remarks (Optional)</label>
                <input
                  type="text"
                  value={formData.remarks}
                  onChange={(e) => setFormData({ ...formData, remarks: e.target.value })}
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                  placeholder="Payment for..."
                />
              </div>

              {error && (
                <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg">
                  {error}
                </div>
              )}

              {result && result.success && (
                <div className="bg-green-50 border border-green-200 text-green-700 px-4 py-3 rounded-lg">
                  <div className="flex items-center space-x-2">
                    <CheckCircle size={20} />
                    <span className="font-semibold">{result.message}</span>
                  </div>
                  {result.transaction_id && (
                    <p className="text-sm mt-1">Transaction ID: {result.transaction_id}</p>
                  )}
                </div>
              )}

              <button
                type="submit"
                disabled={loading}
                className="w-full bg-primary-600 text-white py-3 rounded-lg font-semibold hover:bg-primary-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center space-x-2"
              >
                {loading ? (
                  <>
                    <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-white"></div>
                    <span>Processing...</span>
                  </>
                ) : (
                  <>
                    <ArrowRightLeft size={20} />
                    <span>Transfer Funds</span>
                  </>
                )}
              </button>
            </form>
          </div>
        </div>

        <div className="bg-white rounded-xl shadow-md p-6">
          <h3 className="text-lg font-semibold text-gray-800 mb-4">Transfer Information</h3>
          <div className="space-y-3 text-sm">
            <div>
              <p className="text-gray-600">NEFT</p>
              <p className="text-gray-500">Available 24/7, processed in batches</p>
            </div>
            <div>
              <p className="text-gray-600">RTGS</p>
              <p className="text-gray-500">Real-time, minimum ₹2 lakhs</p>
            </div>
            <div>
              <p className="text-gray-600">IMPS</p>
              <p className="text-gray-500">Instant transfer, 24/7</p>
            </div>
            <div>
              <p className="text-gray-600">UPI</p>
              <p className="text-gray-500">Instant, using UPI ID</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

