# Web UI Setup Guide

## ✅ Web UI Created Successfully!

A complete, modern web interface has been created in the `web-ui/` folder.

## 🎨 Features

- **Modern Design** - Beautiful, responsive UI with Tailwind CSS
- **Dashboard** - Overview with balance and recent transactions
- **Balance View** - Detailed account balance information
- **Fund Transfer** - NEFT, RTGS, IMPS, UPI transfers with AI orchestration
- **Transaction History** - View statements and transaction history
- **Beneficiary Management** - Add and manage beneficiaries
- **AI Assistant** - Natural language banking operations

## 📁 Project Structure

```
web-ui/
├── src/
│   ├── components/
│   │   └── Layout.jsx          # Main layout with sidebar
│   ├── pages/
│   │   ├── Dashboard.jsx       # Dashboard page
│   │   ├── Balance.jsx         # Balance page
│   │   ├── Transfer.jsx         # Fund transfer page
│   │   ├── Statement.jsx        # Statement page
│   │   ├── Beneficiaries.jsx    # Beneficiaries page
│   │   └── AIAssistant.jsx      # AI chat assistant
│   ├── services/
│   │   └── api.js              # API integration layer
│   ├── context/
│   │   └── AuthContext.jsx     # User context
│   ├── App.jsx                 # Main app component
│   ├── main.jsx                # Entry point
│   └── index.css               # Global styles
├── package.json                # Dependencies
├── vite.config.js              # Vite configuration
├── tailwind.config.js          # Tailwind configuration
└── README.md                   # Documentation
```

## 🚀 Quick Start

### 1. Install Dependencies

```bash
cd web-ui
npm install
```

### 2. Configure Environment

Create `.env` file:
```env
VITE_API_BASE_URL=http://localhost:8081
VITE_API_KEY=test-api-key
```

### 3. Start Development Server

```bash
npm run dev
```

The UI will be available at `http://localhost:3000`

## 🔌 Backend Integration

The UI integrates with all backend layers:

1. **Layer 2: AI Skin Orchestrator** (Port 8081)
   - Natural language processing
   - Intent recognition
   - Context enrichment

2. **Layer 1: MCP Server** (Port 8080)
   - Task orchestration
   - Agent coordination
   - Task result retrieval

3. **Layer 5: Banking Integrations** (Port 7000)
   - Balance inquiries
   - Fund transfers
   - Transaction history
   - Beneficiary management

## 📱 Pages

- **Dashboard** (`/dashboard`) - Overview and quick stats
- **Balance** (`/balance`) - Account balance details
- **Transfer** (`/transfer`) - Fund transfer interface
- **Statement** (`/statement`) - Transaction history
- **Beneficiaries** (`/beneficiaries`) - Manage saved beneficiaries
- **AI Assistant** (`/ai-assistant`) - Chat interface for banking operations

## 🎯 Key Features

### AI Assistant
- Natural language input
- Processes requests through AI Skin Orchestrator
- Real-time responses
- Risk score display

### Fund Transfer
- Multiple transfer types (NEFT, RTGS, IMPS, UPI)
- AI-powered orchestration
- Real-time status updates
- Transaction confirmation

### Responsive Design
- Mobile-friendly
- Tablet optimized
- Desktop experience
- Sidebar navigation

## 🛠️ Technology Stack

- **React 18** - UI framework
- **Vite** - Build tool and dev server
- **React Router** - Client-side routing
- **Tailwind CSS** - Utility-first CSS
- **Axios** - HTTP client
- **Lucide React** - Icon library

## 📦 Build for Production

```bash
npm run build
```

Output will be in `dist/` directory.

## 🔧 Development

- Hot module replacement enabled
- API proxy configured for CORS
- ESLint for code quality
- Tailwind CSS for styling

## 📝 Notes

- All API calls are configured in `src/services/api.js`
- User context is managed in `src/context/AuthContext.jsx`
- Default user: U10001 (John Doe)
- All backend services must be running for full functionality

## 🎉 Ready to Use!

The web UI is fully integrated with your backend and ready to use. Just start the backend services and the frontend, and you're good to go!

