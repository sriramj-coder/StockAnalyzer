# 📈 StockAnalyzer – Go + React Trading Dashboard

StockAnalyzer is a full-stack trading dashboard that displays live stock data with technical indicators such as RSI, SMA, EMA, Bollinger Bands, and MACD.  
It uses the Alpaca API (free IEX feed) for data and a modern React + Vite + Tailwind frontend for interactive charts.

## 🚀 Features

### Backend (Go)
- Fetches OHLCV stock data from Alpaca
- Calculates technical indicators: RSI, SMA, EMA, Bollinger Bands, MACD
- Provides a REST API for the frontend

### Frontend (React + Recharts + Tailwind)
- Price chart with Bollinger Bands, SMA, EMA
- RSI chart (0–100 scale)
- MACD chart with signal line and histogram



## 📊 Visualizations

StockAnalyzer provides interactive charts that display live stock prices along with technical indicators.

### Price + Indicators
<img width="974" height="697" alt="image" src="https://github.com/user-attachments/assets/982a5d57-36ec-481c-a0b3-c741c4710789" />

*Shows stock closing prices with Bollinger Bands, SMA, and EMA.*

### RSI (Relative Strength Index)
<img width="917" height="374" alt="image" src="https://github.com/user-attachments/assets/dccaeb3f-bc80-4960-b03e-d792bc562163" />

*RSI chart (0–100 scale) indicating overbought and oversold conditions.*

### MACD (Moving Average Convergence Divergence)
<img width="895" height="452" alt="image" src="https://github.com/user-attachments/assets/a5274d08-a050-485d-bd8a-24e3321b99d9" />

*MACD line, signal line, and histogram to track trend momentum and reversals.*
