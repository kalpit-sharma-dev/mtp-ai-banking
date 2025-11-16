import axios from 'axios'

// Use proxy in development (via Vite), direct URL in production
// In dev: use '/api' which Vite proxies to http://localhost:8081/api/v1
// In prod: use full URL
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || (import.meta.env.DEV ? '/api' : 'http://localhost:8081')
const API_KEY = import.meta.env.VITE_API_KEY || 'test-api-key'

// Add request interceptor for error handling
const setupAxiosInstance = (baseURL) => {
  const instance = axios.create({
    baseURL: baseURL,
    headers: {
      'Content-Type': 'application/json',
      'X-API-Key': API_KEY,
    },
  })

  // Add response interceptor for error handling
  instance.interceptors.response.use(
    (response) => response,
    (error) => {
      // Log full error for debugging
      console.error('API Error Details:', {
        message: error.message,
        response: error.response?.data,
        status: error.response?.status,
        request: error.request,
        config: error.config,
      })

      if (error.response) {
        // Server responded with error
        return Promise.reject({
          message: error.response.data?.error || error.response.data?.message || 'An error occurred',
          status: error.response.status,
          data: error.response.data,
        })
      } else if (error.request) {
        // Request made but no response - check if it's a CORS issue
        const url = error.config?.url || error.config?.baseURL || 'unknown'
        const isCorsError = error.message?.includes('CORS') || error.message?.includes('Network Error')
        
        return Promise.reject({
          message: isCorsError 
            ? 'CORS error. Please check if the service allows cross-origin requests.'
            : `Network error connecting to ${url}. Please check if the backend service is running.`,
          status: 0,
          isNetworkError: true,
        })
      } else {
        // Something else happened
        return Promise.reject({
          message: error.message || 'An unexpected error occurred',
          status: 0,
        })
      }
    }
  )

  return instance
}

const api = setupAxiosInstance(API_BASE_URL)

// AI Skin Orchestrator API
export const orchestratorAPI = {
  // Process natural language request
  processRequest: async (userInput, userId, channel = 'MB', sessionId = null) => {
    const payload = {
      user_id: userId,
      channel: channel,
      input: userInput,
      input_type: 'natural_language',
    }
    if (sessionId) {
      payload.session_id = sessionId
    }
    // In dev: '/api/process' -> Vite proxy -> 'http://localhost:8081/api/v1/process'
    // In prod: 'http://localhost:8081/api/v1/process'
    const endpoint = import.meta.env.DEV ? '/process' : '/api/v1/process'
    const response = await api.post(endpoint, payload)
    return response.data
  },
}

// MCP Server API (direct)
const mcpAPI = setupAxiosInstance('http://localhost:8080')

export const mcpServerAPI = {
  // Submit task
  submitTask: async (taskData) => {
    const response = await mcpAPI.post('/api/v1/submit-task', taskData)
    return response.data
  },

  // Get task result
  getTaskResult: async (taskId) => {
    const response = await mcpAPI.get(`/api/v1/get-result/${taskId}`)
    return response.data
  },

  // Create session
  createSession: async (userId, channel) => {
    const response = await mcpAPI.post('/api/v1/create-session', {
      user_id: userId,
      channel: channel,
    })
    return response.data
  },
}

// Banking Integrations API
// In dev: use '/banking' which Vite proxies to http://localhost:7000
// In prod: use full URL
const BANKING_API_BASE_URL = import.meta.env.VITE_BANKING_API_BASE_URL || (import.meta.env.DEV ? '/banking' : 'http://localhost:7000')
const bankingAPIClient = setupAxiosInstance(BANKING_API_BASE_URL)

export const bankingAPI = {
  // Get balance
  getBalance: async (userId, accountId, channel = 'MB') => {
    const response = await bankingAPIClient.post('/api/v1/balance', {
      user_id: userId,
      account_id: accountId,
      channel: channel,
    })
    return response.data
  },

  // Transfer funds
  transferFunds: async (transferData) => {
    const response = await bankingAPIClient.post('/api/v1/transfer', transferData)
    return response.data
  },

  // Get statement
  getStatement: async (userId, accountId, fromDate, toDate, channel = 'MB') => {
    const response = await bankingAPIClient.post('/api/v1/statement', {
      user_id: userId,
      account_id: accountId,
      start_date: fromDate,
      end_date: toDate,
      channel: channel,
    })
    return response.data
  },

  // Get transaction history
  getTransactionHistory: async (userId, days = 90) => {
    const response = await bankingAPIClient.get(`/api/v1/dwh/history/${userId}?days=${days}`)
    return response.data
  },

  // Add beneficiary
  addBeneficiary: async (beneficiaryData) => {
    const response = await bankingAPIClient.post('/api/v1/beneficiary', beneficiaryData)
    return response.data
  },

  // Get beneficiaries (using DWH query)
  getBeneficiaries: async (userId) => {
    // Use DWH query to get beneficiaries
    try {
      const response = await bankingAPIClient.post('/api/v1/dwh/query', {
        query_type: 'BENEFICIARIES',
        user_id: userId,
      })
      // DWH returns { query_type, data: [...], count, executed_at }
      // Extract beneficiaries from data array
      const beneficiaries = response.data?.data || []
      return { beneficiaries }
    } catch (error) {
      // If error, return empty array
      console.warn('Failed to get beneficiaries:', error)
      return { beneficiaries: [] }
    }
  },
}

export default api

