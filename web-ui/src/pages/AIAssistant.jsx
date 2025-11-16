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
      isTyping: false,
    },
  ])
  const [input, setInput] = useState('')
  const [loading, setLoading] = useState(false)
  const [sessionId, setSessionId] = useState('')
  const messagesEndRef = useRef(null)
  const typingTimeoutRef = useRef(null)

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }

  useEffect(() => {
    scrollToBottom()
  }, [messages])

  // Typing effect function
  const typeMessage = (messageObj, fullText, messageIndex) => {
    let currentIndex = 0
    const typingSpeed = 20 // milliseconds per character

    const typeChar = () => {
      if (currentIndex < fullText.length) {
        const newContent = fullText.substring(0, currentIndex + 1)
        setMessages((prev) => {
          const updated = [...prev]
          updated[messageIndex] = {
            ...messageObj,
            content: newContent,
            isTyping: currentIndex < fullText.length - 1,
          }
          return updated
        })
        currentIndex++
        typingTimeoutRef.current = setTimeout(typeChar, typingSpeed)
      } else {
        // Typing complete
        setMessages((prev) => {
          const updated = [...prev]
          updated[messageIndex] = {
            ...messageObj,
            content: fullText,
            isTyping: false,
          }
          return updated
        })
      }
    }

    typeChar()
  }

  // Cleanup typing timeout on unmount
  useEffect(() => {
    return () => {
      if (typingTimeoutRef.current) {
        clearTimeout(typingTimeoutRef.current)
      }
    }
  }, [])

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
      
      // Process request through AI Skin Orchestrator (with session ID)
      const response = await orchestratorAPI.processRequest(input, user.id, user.channel, sessionId)
      
      console.log('Received response from orchestrator:', response)

      // Build response message from available fields
      let message = ''
      const finalResult = response.final_result || {}
      
      // Extract and format balance information
      if (finalResult.balance !== undefined) {
        const balance = finalResult.balance
        const currency = finalResult.currency || 'INR'
        const accountId = finalResult.account_id || finalResult.account_number || ''
        const formattedBalance = new Intl.NumberFormat('en-IN', {
          style: 'currency',
          currency: currency,
          minimumFractionDigits: 2,
        }).format(balance)
        
        // Show available balance if it exists and is different from total balance
        const availableBalance = finalResult.available_balance
        const hasAvailableBalance = availableBalance !== undefined && availableBalance !== balance
        
        if (hasAvailableBalance) {
          // Show both balances with clear labels
          const formattedAvailable = new Intl.NumberFormat('en-IN', {
            style: 'currency',
            currency: currency,
            minimumFractionDigits: 2,
          }).format(availableBalance)
          
          message = `**Account Balance:** ${formattedBalance}`
          if (accountId) {
            message += `\nAccount: ${accountId}`
          }
          message += `\n\n**Available Balance:** ${formattedAvailable}`
          message += `\n\n*Available balance is the amount you can use (excludes holds and pending transactions)*`
        } else {
          // Show only total balance if available balance is same or not provided
          message = `Your account balance is **${formattedBalance}**`
          if (accountId) {
            message += ` (Account: ${accountId})`
          }
        }
      }
      // Extract conversational messages (greetings, capability questions, etc.)
      else if (finalResult.type === 'conversational' || finalResult.message) {
        message = finalResult.message || response.explanation || 'I\'m here to help you with your banking needs!'
      }
      // Extract statement/transactions
      else if (finalResult.transactions && Array.isArray(finalResult.transactions)) {
        const count = finalResult.transactions.length
        message = `Found ${count} transaction${count !== 1 ? 's' : ''}.\n\n`
        if (count > 0) {
          message += 'Recent transactions:\n'
          finalResult.transactions.slice(0, 5).forEach((txn, idx) => {
            const amount = new Intl.NumberFormat('en-IN', {
              style: 'currency',
              currency: txn.currency || 'INR',
              minimumFractionDigits: 2,
            }).format(txn.amount || 0)
            const date = txn.date || txn.timestamp || txn.created_at || 'N/A'
            message += `${idx + 1}. ${txn.description || 'Transaction'}: ${amount} (${date})\n`
          })
          if (count > 5) {
            message += `\n... and ${count - 5} more transaction${count - 5 !== 1 ? 's' : ''}`
          }
        }
      }
      // Extract beneficiary information
      else if (finalResult.beneficiary_id || finalResult.beneficiaries || finalResult.account_number) {
        if (finalResult.beneficiaries && Array.isArray(finalResult.beneficiaries)) {
          const count = finalResult.beneficiaries.length
          message = `You have ${count} saved beneficiary${count !== 1 ? 'ies' : ''}.\n\n`
          finalResult.beneficiaries.forEach((ben, idx) => {
            message += `${idx + 1}. ${ben.name || 'Beneficiary'} - ${ben.account_number || ''} (${ben.ifsc || ''})\n`
          })
        } else if (finalResult.beneficiary_id || finalResult.account_number) {
          // Single beneficiary addition
          const name = finalResult.name || 'Beneficiary'
          const account = finalResult.account_number || finalResult.account || ''
          const ifsc = finalResult.ifsc || ''
          const beneficiaryId = finalResult.beneficiary_id || ''
          
          message = finalResult.message || `✅ Beneficiary added successfully!\n\n`
          message += `Name: ${name}\n`
          message += `Account Number: ${account}\n`
          if (ifsc) message += `IFSC: ${ifsc}\n`
          if (beneficiaryId) message += `Beneficiary ID: ${beneficiaryId}\n`
          if (finalResult.bank_name) message += `Bank: ${finalResult.bank_name}\n`
          
          // Trigger refresh event for Beneficiaries page
          window.dispatchEvent(new CustomEvent('beneficiaryAdded', { 
            detail: { beneficiaryId, name, account, ifsc } 
          }))
        } else {
          message = finalResult.message || 'Beneficiary operation completed successfully.'
        }
      }
      // Extract transfer/transaction information
      else if (finalResult.transaction_id) {
        const amount = finalResult.amount
        const currency = finalResult.currency || 'INR'
        const formattedAmount = new Intl.NumberFormat('en-IN', {
          style: 'currency',
          currency: currency,
          minimumFractionDigits: 2,
        }).format(amount)
        
        message = `Transaction completed successfully!\n\n`
        message += `Transaction ID: ${finalResult.transaction_id}\n`
        message += `Amount: ${formattedAmount}\n`
        if (finalResult.reference_number) {
          message += `Reference: ${finalResult.reference_number}\n`
        }
        if (finalResult.to_account) {
          message += `To Account: ${finalResult.to_account}\n`
        }
        if (response.explanation) {
          message += `\n${response.explanation}`
        }
        
        // Trigger refresh events for other pages
        window.dispatchEvent(new CustomEvent('transactionCompleted', { 
          detail: { 
            transactionId: finalResult.transaction_id,
            amount,
            toAccount: finalResult.to_account 
          } 
        }))
        window.dispatchEvent(new CustomEvent('balanceUpdated'))
      }
      // Use explanation or message if available
      else {
        message = response.explanation || response.message || finalResult.message || finalResult.explanation
        
        // If still no message, construct from status
        if (!message) {
          const status = response.status || finalResult.status
          if (status === 'APPROVED' || status === 'COMPLETED') {
            message = 'Request processed successfully.'
          } else if (status === 'REJECTED') {
            message = 'Request was rejected. Please check the details and try again.'
          } else {
            message = 'Request processed. Status: ' + (status || 'PENDING')
          }
        }
      }

      // Store session ID for future requests
      if (response.final_result?.session_id) {
        setSessionId(response.final_result.session_id)
      }

      // Add bot message with typing effect
      const botMessage = {
        role: 'bot',
        content: '',
        fullContent: message,
        timestamp: new Date(),
        data: response,
        isTyping: true,
      }

      setMessages((prev) => {
        const newMessages = [...prev, botMessage]
        // Start typing effect after state update
        setTimeout(() => {
          typeMessage(botMessage, message, newMessages.length - 1)
        }, 0)
        return newMessages
      })
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
                <div className="whitespace-pre-wrap">
                  {message.content.split('**').map((part, idx) => 
                    idx % 2 === 1 ? (
                      <span key={idx} className="font-bold text-lg text-primary-600">
                        {part}
                      </span>
                    ) : (
                      <span key={idx}>{part}</span>
                    )
                  )}
                  {message.isTyping && (
                    <span className="inline-block w-2 h-4 bg-gray-600 ml-1 animate-pulse">|</span>
                  )}
                </div>
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
              className="flex-1 px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent text-gray-900 bg-white"
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

