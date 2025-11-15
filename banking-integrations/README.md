# Layer 5: Banking Integrations

The Banking Integrations service provides unified access to Mobile Banking (MB), Net Banking (NB), and Data Warehouse (DWH) systems. This layer acts as the integration gateway for all banking operations.

## Components

### 1. Mobile Banking (MB) Service
Handles mobile banking operations:
- Balance inquiries
- Fund transfers
- Account statements
- Beneficiary management

### 2. Net Banking (NB) Service
Handles net banking operations:
- Balance inquiries
- Fund transfers
- Account statements
- Beneficiary management

### 3. Data Warehouse (DWH) Service
Provides access to data warehouse:
- Transaction history queries
- User profile queries
- Analytics queries
- Historical data retrieval

### 4. Banking Gateway
Unified gateway that routes requests to appropriate channel services.

## Architecture

```
AI Banking Platform
        |
        v
Banking Integrations (Layer 5)
  ├── Mobile Banking (MB) Service
  ├── Net Banking (NB) Service
  └── Data Warehouse (DWH) Service
        |
        v
Core Banking Systems / Databases
```

## Features

✅ **Unified API** - Single API for all banking channels  
✅ **Channel Routing** - Automatically routes to MB/NB based on channel  
✅ **DWH Integration** - Direct access to data warehouse  
✅ **Transaction History** - Retrieves historical transactions  
✅ **Beneficiary Management** - Add and manage beneficiaries  
✅ **Statement Generation** - Account statement retrieval  
✅ **Mock Implementation** - Works without database connections  

## Installation

1. Navigate to the banking-integrations directory:
```bash
cd banking-integrations
```

2. Install dependencies:
```bash
go mod download
```

3. Copy environment file:
```bash
cp .env.example .env
```

4. Configure database connections (optional):
```bash
# Edit .env to enable database connections
DB_ENABLED=true
DB_HOST=localhost
DB_PORT=5432
# ... etc
```

5. Run the service:
```bash
go run cmd/server/main.go
```

The service will start on `http://localhost:7000`

## API Endpoints

### Balance Inquiry

**POST** `/api/v1/balance`

Get account balance.

**Request:**
```json
{
  "user_id": "U10001",
  "account_id": "ACC_001",
  "channel": "MB"
}
```

**Response:**
```json
{
  "account_id": "ACC_001",
  "account_number": "XXXX1234",
  "balance": 150000.0,
  "currency": "INR",
  "available_balance": 145000.0,
  "last_updated": "2024-01-15T10:30:00Z"
}
```

### Fund Transfer

**POST** `/api/v1/transfer`

Transfer funds.

**Request:**
```json
{
  "user_id": "U10001",
  "from_account": "XXXX1234",
  "to_account": "YYYY5678",
  "ifsc": "BANK0001234",
  "amount": 50000,
  "type": "NEFT",
  "channel": "MB",
  "remarks": "Payment for services"
}
```

**Response:**
```json
{
  "transaction_id": "MB_abc12345",
  "status": "COMPLETED",
  "amount": 50000,
  "from_account": "XXXX1234",
  "to_account": "YYYY5678",
  "reference_number": "REFxyz789012",
  "processed_at": "2024-01-15T10:30:00Z",
  "message": "Transfer processed successfully"
}
```

### Account Statement

**POST** `/api/v1/statement`

Get account statement.

**Request:**
```json
{
  "account_id": "ACC_001",
  "user_id": "U10001",
  "start_date": "2024-01-01T00:00:00Z",
  "end_date": "2024-01-31T23:59:59Z",
  "channel": "MB",
  "limit": 100
}
```

### Add Beneficiary

**POST** `/api/v1/beneficiary`

Add a new beneficiary.

**Request:**
```json
{
  "user_id": "U10001",
  "account_number": "YYYY5678",
  "ifsc": "BANK0001234",
  "name": "John Doe",
  "channel": "MB"
}
```

### DWH Query

**POST** `/api/v1/dwh/query`

Query data warehouse.

**Request:**
```json
{
  "query_type": "TRANSACTION_HISTORY",
  "user_id": "U10001",
  "start_date": "2024-01-01T00:00:00Z",
  "end_date": "2024-01-31T23:59:59Z",
  "limit": 100
}
```

### Transaction History

**GET** `/api/v1/dwh/history/{userID}?days=90`

Get transaction history for a user.

## Integration with Other Layers

### Layer 2 (AI Skin Orchestrator)
The orchestrator can call this service to:
- Retrieve user transaction history
- Get account balances
- Process fund transfers

### Layer 3 (Agent Mesh)
Agents can call this service to:
- Banking Agent: Get balances, process transfers
- Fraud Agent: Retrieve transaction history for analysis
- Scoring Agent: Get user profile data

### Layer 4 (ML Models)
ML models can use data from this service for:
- Feature engineering
- Training data
- Real-time predictions

## Configuration

### Environment Variables

- **SERVER_PORT**: Server port (default: 7000)
- **DB_ENABLED**: Enable database connection (default: false)
- **DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME**: Database connection
- **DWH_ENABLED**: Enable DWH connection (default: false)
- **DWH_HOST, DWH_PORT, DWH_USER, DWH_PASSWORD, DWH_NAME**: DWH connection

## Production Considerations

1. **Database Connections**: Connect to actual banking databases
2. **Core Banking Integration**: Integrate with core banking systems
3. **Security**: Implement proper authentication and authorization
4. **Encryption**: Encrypt sensitive data in transit and at rest
5. **Audit Logging**: Log all banking operations
6. **Compliance**: Ensure RBI compliance and regulations
7. **High Availability**: Deploy with redundancy and failover

## Next Steps

This is **Layer 5: Banking Integrations**. The system now has:

- ✅ Layer 1: MCP Server (Orchestration)
- ✅ Layer 2: AI Skin Orchestrator (Intelligence)
- ✅ Layer 3: Agent Mesh (Execution)
- ✅ Layer 4: ML Models (Machine Learning)
- ✅ Layer 5: Banking Integrations (Data Access)

## License

[Your License Here]

