import { useState, useEffect } from 'react'
import { useAuth } from '../context/AuthContext'
import { bankingAPI } from '../services/api'
import { Users, Plus, CheckCircle } from 'lucide-react'

export default function Beneficiaries() {
  const { user } = useAuth()
  const [beneficiaries, setBeneficiaries] = useState([])
  const [loading, setLoading] = useState(true)
  const [showAddForm, setShowAddForm] = useState(false)
  const [formData, setFormData] = useState({
    account_number: '',
    ifsc: '',
    name: '',
  })
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    loadBeneficiaries()
  }, [])

  const loadBeneficiaries = async () => {
    try {
      setLoading(true)
      const data = await bankingAPI.getBeneficiaries(user.id)
      setBeneficiaries(data.beneficiaries || [])
    } catch (error) {
      console.error('Failed to load beneficiaries:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleAddBeneficiary = async (e) => {
    e.preventDefault()
    setSubmitting(true)

    try {
      await bankingAPI.addBeneficiary({
        user_id: user.id,
        account_number: formData.account_number,
        ifsc: formData.ifsc,
        name: formData.name,
        channel: user.channel,
      })
      setShowAddForm(false)
      setFormData({ account_number: '', ifsc: '', name: '' })
      loadBeneficiaries()
    } catch (error) {
      console.error('Failed to add beneficiary:', error)
    } finally {
      setSubmitting(false)
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
          <h1 className="text-3xl font-bold text-gray-800">Beneficiaries</h1>
          <p className="text-gray-600 mt-2">Manage your saved beneficiaries</p>
        </div>
        <button
          onClick={() => setShowAddForm(!showAddForm)}
          className="flex items-center space-x-2 px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition-colors"
        >
          <Plus size={20} />
          <span>Add Beneficiary</span>
        </button>
      </div>

      {showAddForm && (
        <div className="bg-white rounded-xl shadow-md p-6">
          <h3 className="text-lg font-semibold text-gray-800 mb-4">Add New Beneficiary</h3>
          <form onSubmit={handleAddBeneficiary} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">Beneficiary Name</label>
              <input
                type="text"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                required
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">Account Number</label>
              <input
                type="text"
                value={formData.account_number}
                onChange={(e) => setFormData({ ...formData, account_number: e.target.value })}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
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
                required
              />
            </div>
            <div className="flex space-x-2">
              <button
                type="submit"
                disabled={submitting}
                className="flex-1 bg-primary-600 text-white py-2 rounded-lg font-semibold hover:bg-primary-700 transition-colors disabled:opacity-50"
              >
                {submitting ? 'Adding...' : 'Add Beneficiary'}
              </button>
              <button
                type="button"
                onClick={() => {
                  setShowAddForm(false)
                  setFormData({ account_number: '', ifsc: '', name: '' })
                }}
                className="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
              >
                Cancel
              </button>
            </div>
          </form>
        </div>
      )}

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {beneficiaries.length > 0 ? (
          beneficiaries.map((beneficiary, index) => (
            <div key={index} className="bg-white rounded-xl shadow-md p-6">
              <div className="flex items-start justify-between mb-4">
                <div className="flex items-center space-x-3">
                  <div className="w-12 h-12 bg-primary-100 rounded-full flex items-center justify-center">
                    <Users className="text-primary-600" size={24} />
                  </div>
                  <div>
                    <h3 className="font-semibold text-gray-800">{beneficiary.name}</h3>
                    <p className="text-sm text-gray-500">{beneficiary.account_number}</p>
                  </div>
                </div>
                {beneficiary.verified && (
                  <CheckCircle className="text-green-500" size={20} />
                )}
              </div>
              <div className="space-y-2 text-sm">
                <div className="flex justify-between">
                  <span className="text-gray-600">IFSC</span>
                  <span className="font-semibold">{beneficiary.ifsc}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600">Bank</span>
                  <span className="font-semibold">{beneficiary.bank_name || 'N/A'}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600">Status</span>
                  <span
                    className={`font-semibold ${
                      beneficiary.status === 'ACTIVE' ? 'text-green-600' : 'text-gray-600'
                    }`}
                  >
                    {beneficiary.status || 'ACTIVE'}
                  </span>
                </div>
              </div>
            </div>
          ))
        ) : (
          <div className="col-span-full text-center py-12">
            <Users className="mx-auto text-gray-400 mb-4" size={48} />
            <p className="text-gray-500">No beneficiaries added yet</p>
            <p className="text-sm text-gray-400 mt-2">Add a beneficiary to get started</p>
          </div>
        )}
      </div>
    </div>
  )
}

