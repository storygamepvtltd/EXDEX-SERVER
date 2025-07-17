📝 Spot Trading API
Base Paths
bash
Copy
Edit
POST /api/v1/spot/market    → Place Market Order
POST /api/v1/spot/limit     → Place Limit Order
All endpoints require authentication via JWT (attached with Authorization: Bearer <token> header).

1️⃣ Market Order
Description: Execute a buy or sell order at the current market price.

Endpoint:

bash
Copy
Edit
POST /api/v1/spot/market
Headers:

pgsql
Copy
Edit
Authorization: Bearer <token>
Content-Type: application/json
Request Body:

json
Copy
Edit
{
  "symbol": "BTCUSDT",
  "side": "BUY",         // or "SELL"
  "quantity": "0.001"
}
symbol: Trading pair (e.g. BTCUSDT)

side: "BUY" or "SELL"

quantity: Amount of base asset to trade

Successful Response (HTTP 200):

json
Copy
Edit
{
  "symbol": "BTCUSDT",
  "orderId": 12345678,
  "transactTime": 1625239123456,
  "price": "0.00000000",
  "origQty": "0.00100000",
  "executedQty": "0.00100000",
  "status": "FILLED",
  "type": "MARKET",
  "side": "BUY"
}
Error Responses:

HTTP 400 – Missing or invalid input

HTTP 401 – Unauthorized / invalid JWT

HTTP 500 – Binance API or backend error

2️⃣ Limit Order
Description: Place a buy or sell order at a specified price; remains open until executed or canceled.

Endpoint:

bash
Copy
Edit
POST /api/v1/spot/limit
Headers:

pgsql
Copy
Edit
Authorization: Bearer <token>
Content-Type: application/json
Request Body:

json
Copy
Edit
{
  "symbol": "BTCUSDT",
  "side": "BUY",         // or "SELL"
  "quantity": "0.001",
  "price": "60000.00",   // Desired price per unit
  "timeInForce": "GTC"   // GTC / IOC / FOK
}
Successful Response (HTTP 200):

json
Copy
Edit
{
  "symbol": "BTCUSDT",
  "orderId": 87654321,
  "clientOrderId": "abc123xyz",
  "transactTime": 1625239123456,
  "price": "60000.00",
  "origQty": "0.00100000",
  "executedQty": "0.00000000",
  "status": "NEW",
  "type": "LIMIT",
  "side": "BUY",
  "timeInForce": "GTC"
}
Error Responses:

HTTP 400 – Bad request parameters

HTTP 401 – Unauthorized / invalid JWT

HTTP 500 – Binance API or backend error

Authentication
All routes use JWT verification.
The middleware inspects:

Authorization header must include Bearer <token>

Token must be valid and unexpired

Extracted claims: userID, email, role

Unauthorized requests receive:

json
Copy
Edit
{
  "status": false,
  "error": "Invalid or expired token"
}
Common Errors
400 Bad Request: Returned for missing/invalid JSON input

401 Unauthorized: Invalid or missing JWT

500 Internal Server Error:

Binance failed

Duplicate records

Other back-end issues

Example Usage
Market Order via curl:
bash
Copy
Edit
curl -X POST http://localhost:3000/api/v1/spot/market \
  -H "Authorization: Bearer <jwt_token>" \
  -H "Content-Type: application/json" \
  -d '{"symbol":"BTCUSDT","side":"BUY","quantity":"0.001"}'
Limit Order via curl:
bash
Copy
Edit
curl -X POST http://localhost:3000/api/v1/spot/limit \
  -H "Authorization: Bearer <jwt_token>" \
  -H "Content-Type: application/json" \
  -d '{"symbol":"BTCUSDT","side":"SELL","quantity":"0.002","price":"65000.00","timeInForce":"GTC"}'
📁 Suggested File: ORDER_API_DOC.md
Save the above documentation as ORDER_API_DOC.md in your project root or docs/ folder. It’s ready for sharing with frontend or QA teams for implementation guidance.

Would you like auto-generated API clients (OpenAPI spec) or example Postman collection next?









Ask ChatGPT
