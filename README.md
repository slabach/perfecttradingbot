

# .env
```
.env - Trading Bot Configuration

# Alpaca API Keys
APCA_API_KEY_ID=***
APCA_API_SECRET_KEY=***

# Live trading mode
LIVE_TRADING=false

# Strategy filters
MIN_HISTOGRAM=0.005              # Minimum MACD histogram to trigger trade
REQUIRE_TREND=true               # Require price to be above long EMA to confirm uptrend
LONG_EMA_PERIOD=20               # Long EMA period for trend confirmation
MACD_SLOPE_THRESHOLD=0.005       # Minimum slope (momentum) of MACD line to allow trade
# MIN_PERCENT_MOVE=0.5           # Min % move since last cross to avoid range-chop
USE_RSI_FILTER=true
RSI_MIN=40
RSI_MAX=70

# Portfolio simulation
STARTING_BALANCE=1500.00

# Trade sizing logic
USE_POSITION_PERCENT=true      # If true, use percent of available funds to size trades
TRADE_FUND_PERCENT=0.95        # Use 95% of available cash
ROUND_SHARES=true              # Round down to nearest whole share

# File Path
FILE_DIR=***
```