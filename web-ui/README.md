# AI Banking Platform - Web UI

Modern, responsive web interface for the AI Banking Platform.

## Features

- 🎨 **Modern UI** - Beautiful, responsive design with Tailwind CSS
- 💬 **AI Assistant** - Natural language banking operations
- 💰 **Balance Management** - View account balances
- 🔄 **Fund Transfer** - NEFT, RTGS, IMPS, UPI transfers
- 📊 **Transaction History** - View statements and history
- 👥 **Beneficiary Management** - Add and manage beneficiaries
- 📱 **Responsive Design** - Works on desktop, tablet, and mobile

## Tech Stack

- **React 18** - UI framework
- **Vite** - Build tool
- **React Router** - Navigation
- **Tailwind CSS** - Styling
- **Axios** - HTTP client
- **Lucide React** - Icons

## Installation

1. Navigate to the web-ui directory:
```bash
cd web-ui
```

2. Install dependencies:
```bash
npm install
```

3. Create environment file:
```bash
cp .env.example .env
```

4. Start development server:
```bash
npm run dev
```

The app will be available at `http://localhost:3000`

## Environment Variables

Create a `.env` file in the `web-ui` directory:

```env
VITE_API_BASE_URL=http://localhost:8081
VITE_API_KEY=test-api-key
```

## Backend Integration

The UI integrates with:

- **Layer 2: AI Skin Orchestrator** (Port 8081) - Natural language processing
- **Layer 1: MCP Server** (Port 8080) - Task orchestration
- **Layer 5: Banking Integrations** (Port 7000) - Banking operations

## Available Pages

- `/dashboard` - Overview and recent transactions
- `/balance` - Account balance details
- `/transfer` - Fund transfer interface
- `/statement` - Transaction history
- `/beneficiaries` - Manage beneficiaries
- `/ai-assistant` - Chat with AI assistant

## Build for Production

```bash
npm run build
```

The built files will be in the `dist` directory.

## Project Structure

```
web-ui/
├── src/
│   ├── components/     # Reusable components
│   ├── pages/          # Page components
│   ├── services/       # API services
│   ├── context/        # React context
│   ├── App.jsx        # Main app component
│   └── main.jsx       # Entry point
├── public/            # Static assets
├── index.html         # HTML template
└── package.json       # Dependencies
```

## Development

The development server runs on port 3000 with hot module replacement.

API requests are proxied through Vite to avoid CORS issues.

## License

[Your License Here]

