import { useState, useRef, useEffect } from 'react'
import { useAuth } from '../context/AuthContext'
import { orchestratorAPI } from '../services/api'
import { Send, Bot, User, Loader } from 'lucide-react'

export default function AIAssistant() {
  const { user } = useAuth()
  const [messages, setMessages] = useState([
    {
      role: 'bot',
      content: 'Hello! I\'m your AI banking assistant. How can I help you today?',
      timestamp: new Date(),
    },
  ])
  const [input, setInput] = useState('')
  const [loading, setLoading] = useState(false)
  const messagesEndRef = useRef(null)

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }

  useEffect(() => {
    scrollToBottom()
  }, [messages])

  const handleSubmit = async (e) => {
    e.preventDefault()
    if (!input.trim() || loading) return

    const userMessage = {
      role: 'user',
      content: input,
      timestamp: new Date(),
    }

    setMessages((prev) => [...prev, userMessage])
    setInput('')
    setLoading(true)

    try {
      // Validate user data
      if (!user || !user.id || !user.channel) {
        throw new Error('User information is missing. Please refresh the page.')
      }

      console.log('Sending request to orchestrator:', { input, userId: user.id, channel: user.channel })
      
      // Process request through AI Skin Orchestrator
      const response = await orchestratorAPI.processRequest(input, user.id, user.channel)
      
      console.log('Received response from orchestrator:', response)

      // Build response message from available fields
      let message = response.explanation || response.message
      
      // If no explanation, try to get from final_result
      if (!message && response.final_result) {
        message = response.final_result.message || response.final_result.explanation
      }
      
      // If still no message, construct from status
      if (!message) {
        const status = response.status || response.final_result?.status
        if (status === 'APPROVED' || status === 'COMPLETED') {
          message = 'Request processed successfully.'
        } else if (status === 'REJECTED') {
          message = 'Request was rejected. Please check the details and try again.'
        } else {
          message = 'Request processed. Status: ' + (status || 'PENDING')
        }
      }

      const botMessage = {
        role: 'bot',
        content: message,
        timestamp: new Date(),
        data: response,
      }

      setMessages((prev) => [...prev, botMessage])
    } catch (error) {
      // Extract error message from axios interceptor format
      let errorMsg = 'Sorry, I encountered an error processing your request. Please try again.'
      
      if (error.message) {
        errorMsg = error.message
      } else if (error.data?.error) {
        errorMsg = error.data.error
      } else if (error.data?.message) {
        errorMsg = error.data.message
      } else if (typeof error === 'string') {
        errorMsg = error
      }
      
      // Show more detailed error message
      const errorMessage = {
        role: 'bot',
        content: errorMsg,
        timestamp: new Date(),
        error: true,
      }
      setMessages((prev) => [...prev, errorMessage])
      
      // Log full error for debugging
      console.error('AI Assistant error:', error)
      console.error('Error details:', {
        message: error.message,
        status: error.status,
        data: error.data,
        fullError: error
      })
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-gray-800">AI Banking Assistant</h1>
        <p className="text-gray-600 mt-2">Ask me anything about your banking needs</p>
      </div>

      <div className="bg-white rounded-xl shadow-md overflow-hidden flex flex-col" style={{ height: '600px' }}>
        {/* Messages */}
        <div className="flex-1 overflow-y-auto p-6 space-y-4">
          {messages.map((message, index) => (
            <div
              key={index}
              className={`flex items-start space-x-3 ${
                message.role === 'user' ? 'flex-row-reverse space-x-reverse' : ''
              }`}
            >
              <div
                className={`flex-shrink-0 w-10 h-10 rounded-full flex items-center justify-center ${
                  message.role === 'user'
                    ? 'bg-primary-600 text-white'
                    : 'bg-gray-200 text-gray-700'
                }`}
              >
                {message.role === 'user' ? <User size={20} /> : <Bot size={20} />}
              </div>
              <div
                className={`flex-1 rounded-lg p-4 ${
                  message.role === 'user'
                    ? 'bg-primary-600 text-white'
                    : message.error
                    ? 'bg-red-50 text-red-700 border border-red-200'
                    : 'bg-gray-100 text-gray-800'
                }`}
              >
                <p className="whitespace-pre-wrap">{message.content}</p>
                {message.data && message.data.risk_score !== undefined && (
                  <p className="text-xs mt-2 opacity-75">
                    Risk Score: {(message.data.risk_score * 100).toFixed(1)}%
                  </p>
                )}
                <p className="text-xs mt-2 opacity-75">
                  {message.timestamp.toLocaleTimeString()}
                </p>
              </div>
            </div>
          ))}
          {loading && (
            <div className="flex items-start space-x-3">
              <div className="flex-shrink-0 w-10 h-10 rounded-full bg-gray-200 text-gray-700 flex items-center justify-center">
                <Bot size={20} />
              </div>
              <div className="flex-1 rounded-lg p-4 bg-gray-100">
                <Loader className="animate-spin" size={20} />
              </div>
            </div>
          )}
          <div ref={messagesEndRef} />
        </div>

        {/* Input */}
        <form onSubmit={handleSubmit} className="border-t p-4">
          <div className="flex space-x-2">
            <input
              type="text"
              value={input}
              onChange={(e) => setInput(e.target.value)}
              placeholder="Type your message... (e.g., 'Transfer 50000 rupees to account XXXX4321 via NEFT')"
              className="flex-1 px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
              disabled={loading}
            />
            <button
              type="submit"
              disabled={loading || !input.trim()}
              className="bg-primary-600 text-white px-6 py-2 rounded-lg hover:bg-primary-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center space-x-2"
            >
              <Send size={20} />
              <span>Send</span>
            </button>
          </div>
          <p className="text-xs text-gray-500 mt-2">
            Try: "Check my balance", "Transfer money", "Show my statement"
          </p>
        </form>
      </div>
    </div>
  )
}

