# API coverage — 116 operations

Generated from the live schema (https://api.cryptodatadownload.com/?format=openapi). Method names derive from each
operation's path suffix inside its service group; `*Raw` variants exist only on
operations whose schema declares a `return` parameter.


## Aggregates

| Method | Endpoint | Raw | Docs |
|---|---|---|---|
| `Aggregates.FeaturesMlAll(ctx, FeaturesMlAllParams{})` | `GET https://api.cryptodatadownload.com/v1/data/aggregates/features/ml/all/` — All ML Features Composite | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1aggregates~1features~1ml~1all~1/get) |
| `Aggregates.OnchainDashboard(ctx, OnchainDashboardParams{})` | `GET https://api.cryptodatadownload.com/v1/data/aggregates/onchain/dashboard/` — On-Chain Dashboard Composite | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1aggregates~1onchain~1dashboard~1/get) |
| `Aggregates.OptionsDashboard(ctx, OptionsDashboardParams{})` | `GET https://api.cryptodatadownload.com/v1/data/aggregates/options/dashboard/` — Options Dashboard Composite | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1aggregates~1options~1dashboard~1/get) |
| `Aggregates.RiskPanel(ctx, RiskPanelParams{})` | `GET https://api.cryptodatadownload.com/v1/data/aggregates/risk/panel/` — Per-Symbol Risk Panel | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1aggregates~1risk~1panel~1/get) |

## Market Breadth Indicators

| Method | Endpoint | Raw | Docs |
|---|---|---|---|
| `Breadth.N52wkHighsLows(ctx, N52wkHighsLowsParams{})` | `GET https://api.cryptodatadownload.com/v1/data/breadth/52wk-highs-lows/` — Daily New 52 Week Highs and Lows Counts | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1breadth~152wk-highs-lows~1/get) |
| `Breadth.AdvanceDecline(ctx, AdvanceDeclineParams{})` | `GET https://api.cryptodatadownload.com/v1/data/breadth/advance-decline/` — Advance/Decline Calculations | yes | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1breadth~1advance-decline~1/get) |
| `Breadth.MovingAverageTracking(ctx, MovingAverageTrackingParams{})` | `GET https://api.cryptodatadownload.com/v1/data/breadth/moving-average-tracking/` — Pair Counts Trading Above Moving Averages | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1breadth~1moving-average-tracking~1/get) |

## US Govt Data

| Method | Endpoint | Raw | Docs |
|---|---|---|---|
| `Govt.CftcCot(ctx, CftcCotParams{})` | `GET https://api.cryptodatadownload.com/v1/data/cftc/cot/` — Committment of Traders (COT) Timeseries | yes | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1cftc~1cot~1/get) |
| `Govt.TreasuryParYieldsReal(ctx, TreasuryParYieldsRealParams{})` | `GET https://api.cryptodatadownload.com/v1/data/treasury/par-yields-real/` — Treasury Par Yield Curve Real Rates | yes | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1treasury~1par-yields-real~1/get) |
| `Govt.TreasuryParYields(ctx, TreasuryParYieldsParams{})` | `GET https://api.cryptodatadownload.com/v1/data/treasury/par-yields/` — Treasury Par Yield Curve Rates | yes | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1treasury~1par-yields~1/get) |

## Catalog & Coverage

| Method | Endpoint | Raw | Docs |
|---|---|---|---|
| `Catalog.CftcCotAvailable(ctx, CftcCotAvailableParams{})` | `GET https://api.cryptodatadownload.com/v1/data/cftc/cot/available/` — Available COT Contract Names / Groups | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1cftc~1cot~1available~1/get) |
| `Catalog.OhlcBinanceAllAvailable(ctx, OhlcBinanceAllAvailableParams{})` | `GET https://api.cryptodatadownload.com/v1/data/ohlc/binance/all/available/` — Available Binance Symbols | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1ohlc~1binance~1all~1available~1/get) |
| `Catalog.OhlcDeribitFuturesAvailable(ctx, OhlcDeribitFuturesAvailableParams{})` | `GET https://api.cryptodatadownload.com/v1/data/ohlc/deribit/futures/available/` — Available DeriBit Futures OHLC Symbols | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1ohlc~1deribit~1futures~1available~1/get) |
| `Catalog.OhlcDeribitFuturesFundingAvailable(ctx, OhlcDeribitFuturesFundingAvailableParams{})` | `GET https://api.cryptodatadownload.com/v1/data/ohlc/deribit/futures/funding/available/` — Available DeriBit Funding Symbols | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1ohlc~1deribit~1futures~1funding~1available~1/get) |
| `Catalog.OhlcDeribitOptionsAvailable(ctx, OhlcDeribitOptionsAvailableParams{})` | `GET https://api.cryptodatadownload.com/v1/data/ohlc/deribit/options/available/` — Available Deribit Option Symbols | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1ohlc~1deribit~1options~1available~1/get) |
| `Catalog.SummaryBinanceFuturesFundingAvailable(ctx, SummaryBinanceFuturesFundingAvailableParams{})` | `GET https://api.cryptodatadownload.com/v1/data/summary/binance/futures/funding/available/` — Available Binance Futures Funding Symbols | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1binance~1futures~1funding~1available~1/get) |
| `Catalog.SummaryBinanceSpotTransactionalAvailable(ctx, SummaryBinanceSpotTransactionalAvailableParams{})` | `GET https://api.cryptodatadownload.com/v1/data/summary/binance/spot/transactional/available/` — Binance Trade Summary Available Dates | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1binance~1spot~1transactional~1available~1/get) |
| `Catalog.SummaryDeribitOptionsGreeksAvailable(ctx, SummaryDeribitOptionsGreeksAvailableParams{})` | `GET https://api.cryptodatadownload.com/v1/data/summary/deribit/options/greeks/available/` — DeriBit Greeks Available Dates | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1deribit~1options~1greeks~1available~1/get) |
| `Catalog.TransactionsDeribitOptionsAvailable(ctx, TransactionsDeribitOptionsAvailableParams{})` | `GET https://api.cryptodatadownload.com/v1/data/transactions/deribit/options/available/` — Deribit Options Transactions Available | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1transactions~1deribit~1options~1available~1/get) |

## Pre-Computed ML Features — Market Basics

| Method | Endpoint | Raw | Docs |
|---|---|---|---|
| `MLMarket.FuturesBasisTerm(ctx, FuturesBasisTermParams{})` | `GET https://api.cryptodatadownload.com/v1/data/features/futures/basis-term/` — Futures Basis Term Structure | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1futures~1basis-term~1/get) |
| `MLMarket.MacroYieldSpreads(ctx, MacroYieldSpreadsParams{})` | `GET https://api.cryptodatadownload.com/v1/data/features/macro/yield-spreads/` — Macro Yield Spreads | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1macro~1yield-spreads~1/get) |
| `MLMarket.SpotReturns(ctx, SpotReturnsParams{})` | `GET https://api.cryptodatadownload.com/v1/data/features/spot/returns/` — Spot Returns | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1spot~1returns~1/get) |
| `MLMarket.SpotVwap(ctx, SpotVwapParams{})` | `GET https://api.cryptodatadownload.com/v1/data/features/spot/vwap/` — Spot VWAP | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1spot~1vwap~1/get) |

## Pre-Computed ML Features — Futures Positioning

| Method | Endpoint | Raw | Docs |
|---|---|---|---|
| `MLFutures.FundingRegime(ctx, FundingRegimeParams{})` | `GET https://api.cryptodatadownload.com/v1/data/features/futures/funding-regime/` — Funding Regime | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1futures~1funding-regime~1/get) |
| `MLFutures.OiChange(ctx, OiChangeParams{})` | `GET https://api.cryptodatadownload.com/v1/data/features/futures/oi-change/` — Open-Interest Change | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1futures~1oi-change~1/get) |
| `MLFutures.PerpBasisTs(ctx, PerpBasisTsParams{})` | `GET https://api.cryptodatadownload.com/v1/data/features/futures/perp-basis-ts/` — Perp Basis Time-Series | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1futures~1perp-basis-ts~1/get) |

## Pre-Computed ML Features — Macro Signals

| Method | Endpoint | Raw | Docs |
|---|---|---|---|
| `MLMacro.CotPositioning(ctx, CotPositioningParams{})` | `GET https://api.cryptodatadownload.com/v1/data/features/macro/cot-positioning/` — COT Positioning | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1macro~1cot-positioning~1/get) |
| `MLMacro.RealRateRegime(ctx, RealRateRegimeParams{})` | `GET https://api.cryptodatadownload.com/v1/data/features/macro/real-rate-regime/` — Real-Rate Regime | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1macro~1real-rate-regime~1/get) |
| `MLMacro.YieldCurvePca(ctx, YieldCurvePcaParams{})` | `GET https://api.cryptodatadownload.com/v1/data/features/macro/yield-curve-pca/` — Yield-Curve PCA | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1macro~1yield-curve-pca~1/get) |

## Pre-Computed ML Features — Cross-Sectional

| Method | Endpoint | Raw | Docs |
|---|---|---|---|
| `MLCrossSectional.BetaBtc(ctx, BetaBtcParams{})` | `GET https://api.cryptodatadownload.com/v1/data/features/ml/beta-btc/` — Beta vs BTC | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1ml~1beta-btc~1/get) |
| `MLCrossSectional.Momentum(ctx, MomentumParams{})` | `GET https://api.cryptodatadownload.com/v1/data/features/ml/momentum/` — Momentum Factor | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1ml~1momentum~1/get) |
| `MLCrossSectional.ReturnRank(ctx, ReturnRankParams{})` | `GET https://api.cryptodatadownload.com/v1/data/features/ml/return-rank/` — Return Rank | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1ml~1return-rank~1/get) |
| `MLCrossSectional.Reversal(ctx, ReversalParams{})` | `GET https://api.cryptodatadownload.com/v1/data/features/ml/reversal/` — Reversal & Hurst Regime | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1ml~1reversal~1/get) |
| `MLCrossSectional.VolRank(ctx, VolRankParams{})` | `GET https://api.cryptodatadownload.com/v1/data/features/ml/vol-rank/` — Volatility Rank | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1ml~1vol-rank~1/get) |

## Pre-Computed ML Features — Microstructure

| Method | Endpoint | Raw | Docs |
|---|---|---|---|
| `MLMicrostructure.LargeTradeShare(ctx, LargeTradeShareParams{})` | `GET https://api.cryptodatadownload.com/v1/data/features/ml/large-trade-share/` — Large-Trade Share | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1ml~1large-trade-share~1/get) |
| `MLMicrostructure.OrderFlowImbalance(ctx, OrderFlowImbalanceParams{})` | `GET https://api.cryptodatadownload.com/v1/data/features/ml/order-flow-imbalance/` — Order-Flow Imbalance | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1ml~1order-flow-imbalance~1/get) |
| `MLMicrostructure.VolumeProfile(ctx, VolumeProfileParams{})` | `GET https://api.cryptodatadownload.com/v1/data/features/ml/volume-profile/` — Volume Profile | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1ml~1volume-profile~1/get) |
| `MLMicrostructure.XexchangeSpread(ctx, XexchangeSpreadParams{})` | `GET https://api.cryptodatadownload.com/v1/data/features/ml/xexchange-spread/` — Cross-Exchange Spread | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1ml~1xexchange-spread~1/get) |

## On-Chain Data

| Method | Endpoint | Raw | Docs |
|---|---|---|---|
| `OnChain.FeaturesOnchainAddressVelocity(ctx, FeaturesOnchainAddressVelocityParams{})` | `GET https://api.cryptodatadownload.com/v1/data/features/onchain/address-velocity/` — Address Velocity | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1onchain~1address-velocity~1/get) |
| `OnChain.FeaturesOnchainContractGrowth(ctx, FeaturesOnchainContractGrowthParams{})` | `GET https://api.cryptodatadownload.com/v1/data/features/onchain/contract-growth/` — ETH Contract Growth | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1onchain~1contract-growth~1/get) |
| `OnChain.FeaturesOnchainNetworkCongestion(ctx, FeaturesOnchainNetworkCongestionParams{})` | `GET https://api.cryptodatadownload.com/v1/data/features/onchain/network-congestion/` — Network Congestion Score | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1onchain~1network-congestion~1/get) |
| `OnChain.FeaturesOnchainNvt(ctx, FeaturesOnchainNvtParams{})` | `GET https://api.cryptodatadownload.com/v1/data/features/onchain/nvt/` — NVT Ratio | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1onchain~1nvt~1/get) |
| `OnChain.FeaturesOnchainSupplyVelocity(ctx, FeaturesOnchainSupplyVelocityParams{})` | `GET https://api.cryptodatadownload.com/v1/data/features/onchain/supply-velocity/` — Supply Velocity | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1onchain~1supply-velocity~1/get) |
| `OnChain.FeaturesOnchainWhale(ctx, FeaturesOnchainWhaleParams{})` | `GET https://api.cryptodatadownload.com/v1/data/features/onchain/whale/` — Whale Activity | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1onchain~1whale~1/get) |
| `OnChain.OnchainBtcMempool(ctx, OnchainBtcMempoolParams{})` | `GET https://api.cryptodatadownload.com/v1/data/onchain/btc/mempool/` — BTC Mempool State | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1onchain~1btc~1mempool~1/get) |
| `OnChain.OnchainEthExchangeFlows(ctx, OnchainEthExchangeFlowsParams{})` | `GET https://api.cryptodatadownload.com/v1/data/onchain/eth/exchange-flows/` — ETH Exchange Flows | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1onchain~1eth~1exchange-flows~1/get) |
| `OnChain.OnchainEthExchangeReserves(ctx, OnchainEthExchangeReservesParams{})` | `GET https://api.cryptodatadownload.com/v1/data/onchain/eth/exchange-reserves/` — ETH Exchange Reserves | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1onchain~1eth~1exchange-reserves~1/get) |
| `OnChain.OnchainEthGas(ctx, OnchainEthGasParams{})` | `GET https://api.cryptodatadownload.com/v1/data/onchain/eth/gas/` — ETH Gas Prices | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1onchain~1eth~1gas~1/get) |
| `OnChain.OnchainEthTransactions(ctx, OnchainEthTransactionsParams{})` | `GET https://api.cryptodatadownload.com/v1/data/onchain/eth/transactions/` — ETH Transaction Aggregates | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1onchain~1eth~1transactions~1/get) |
| `OnChain.SummaryBitcoinMarketcap(ctx, SummaryBitcoinMarketcapParams{})` | `GET https://api.cryptodatadownload.com/v1/data/summary/bitcoin/marketcap/` — BTC Market Capitalization | yes | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1bitcoin~1marketcap~1/get) |
| `OnChain.SummaryBlockchainBchain(ctx, SummaryBlockchainBchainParams{})` | `GET https://api.cryptodatadownload.com/v1/data/summary/blockchain/bchain/` — BCHAIN Aggregated Data | yes | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1blockchain~1bchain~1/get) |
| `OnChain.SummaryBlockchainBlocks(ctx, SummaryBlockchainBlocksParams{})` | `GET https://api.cryptodatadownload.com/v1/data/summary/blockchain/blocks/` — Blockchain Summary Metrics | yes | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1blockchain~1blocks~1/get) |
| `OnChain.SummaryBlockchainEthContracts(ctx, SummaryBlockchainEthContractsParams{})` | `GET https://api.cryptodatadownload.com/v1/data/summary/blockchain/eth/contracts/` — On-Chain Ethereum (ETH) Summary Contract Creation Metrics | yes | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1blockchain~1eth~1contracts~1/get) |
| `OnChain.SummaryBlockchainTransactions(ctx, SummaryBlockchainTransactionsParams{})` | `GET https://api.cryptodatadownload.com/v1/data/summary/blockchain/transactions/` — Blockchain Transaction Metrics | yes | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1blockchain~1transactions~1/get) |
| `OnChain.SummaryBlockchainTransactionsCoinbase(ctx, SummaryBlockchainTransactionsCoinbaseParams{})` | `GET https://api.cryptodatadownload.com/v1/data/summary/blockchain/transactions/coinbase/` — BTC Coin Creation(s) | yes | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1blockchain~1transactions~1coinbase~1/get) |
| `OnChain.SummaryBlockchainTransactionsLargest(ctx, SummaryBlockchainTransactionsLargestParams{})` | `GET https://api.cryptodatadownload.com/v1/data/summary/blockchain/transactions/largest/` — Top 50 BTC Transactions per Day | yes | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1blockchain~1transactions~1largest~1/get) |

## Pre-Computed ML Features — Options-Implied

| Method | Endpoint | Raw | Docs |
|---|---|---|---|
| `MLOptions.ImpliedMove(ctx, ImpliedMoveParams{})` | `GET https://api.cryptodatadownload.com/v1/data/features/options/implied-move/` — Implied Move | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1options~1implied-move~1/get) |
| `MLOptions.OiConcentration(ctx, OiConcentrationParams{})` | `GET https://api.cryptodatadownload.com/v1/data/features/options/oi-concentration/` — Open-Interest Concentration | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1options~1oi-concentration~1/get) |
| `MLOptions.SkewSignal(ctx, SkewSignalParams{})` | `GET https://api.cryptodatadownload.com/v1/data/features/options/skew-signal/` — Skew Signal | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1options~1skew-signal~1/get) |
| `MLOptions.SmileCurvature(ctx, SmileCurvatureParams{})` | `GET https://api.cryptodatadownload.com/v1/data/features/options/smile-curvature/` — Smile Curvature | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1options~1smile-curvature~1/get) |

## Pre-Computed ML Features — Regime & Anomaly

| Method | Endpoint | Raw | Docs |
|---|---|---|---|
| `MLRegime.AnomalyScore(ctx, AnomalyScoreParams{})` | `GET https://api.cryptodatadownload.com/v1/data/features/regime/anomaly-score/` — Anomaly Score | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1regime~1anomaly-score~1/get) |
| `MLRegime.ClusterMembership(ctx, ClusterMembershipParams{})` | `GET https://api.cryptodatadownload.com/v1/data/features/regime/cluster-membership/` — Cluster Membership | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1regime~1cluster-membership~1/get) |
| `MLRegime.RegimeState(ctx, RegimeStateParams{})` | `GET https://api.cryptodatadownload.com/v1/data/features/regime/regime-state/` — Regime State | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1features~1regime~1regime-state~1/get) |

## Perp Funding Rates

| Method | Endpoint | Raw | Docs |
|---|---|---|---|
| `Funding.FundingAligned(ctx, FundingAlignedParams{})` | `GET https://api.cryptodatadownload.com/v1/data/funding/aligned/` — Aligned Cross-Venue Funding Rates | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1funding~1aligned~1/get) |
| `Funding.FundingBinance(ctx, FundingBinanceParams{})` | `GET https://api.cryptodatadownload.com/v1/data/funding/binance/` — Binance Funding Rates | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1funding~1binance~1/get) |
| `Funding.FundingBybit(ctx, FundingBybitParams{})` | `GET https://api.cryptodatadownload.com/v1/data/funding/bybit/` — Bybit Funding Rates | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1funding~1bybit~1/get) |
| `Funding.FundingOkx(ctx, FundingOkxParams{})` | `GET https://api.cryptodatadownload.com/v1/data/funding/okx/` — OKX Funding Rates | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1funding~1okx~1/get) |
| `Funding.SummaryBinanceFuturesFunding(ctx, SummaryBinanceFuturesFundingParams{})` | `GET https://api.cryptodatadownload.com/v1/data/summary/binance/futures/funding/` — Binance Futures Funding Rates | yes | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1binance~1futures~1funding~1/get) |
| `Funding.SummaryFundingComparison(ctx, SummaryFundingComparisonParams{})` | `GET https://api.cryptodatadownload.com/v1/data/summary/funding/comparison/` — Funding Comparison (Legacy Daily) | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1funding~1comparison~1/get) |

## Binance OHLC

| Method | Endpoint | Raw | Docs |
|---|---|---|---|
| `BinanceOHLC.FuturesCm(ctx, FuturesCmParams{})` | `GET https://api.cryptodatadownload.com/v1/data/ohlc/binance/futures/cm/` — Binance Futures CM Data | yes | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1ohlc~1binance~1futures~1cm~1/get) |
| `BinanceOHLC.FuturesUm(ctx, FuturesUmParams{})` | `GET https://api.cryptodatadownload.com/v1/data/ohlc/binance/futures/um/` — Binance Futures UM Data | yes | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1ohlc~1binance~1futures~1um~1/get) |
| `BinanceOHLC.Options(ctx, OptionsParams{})` | `GET https://api.cryptodatadownload.com/v1/data/ohlc/binance/options/` — Binance Options Data | yes | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1ohlc~1binance~1options~1/get) |
| `BinanceOHLC.Spot(ctx, SpotParams{})` | `GET https://api.cryptodatadownload.com/v1/data/ohlc/binance/spot/` — Binance Spot Data | yes | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1ohlc~1binance~1spot~1/get) |
| `BinanceOHLC.Tick(ctx, TickParams{})` | `GET https://api.cryptodatadownload.com/v1/data/ohlc/binance/tick/` — Binance Tick Data | yes | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1ohlc~1binance~1tick~1/get) |

## DeriBit OHLC

| Method | Endpoint | Raw | Docs |
|---|---|---|---|
| `DeribitOHLC.Futures(ctx, FuturesParams{})` | `GET https://api.cryptodatadownload.com/v1/data/ohlc/deribit/futures/` — DeriBit Futures Daily Data | yes | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1ohlc~1deribit~1futures~1/get) |
| `DeribitOHLC.FuturesFunding(ctx, FuturesFundingParams{})` | `GET https://api.cryptodatadownload.com/v1/data/ohlc/deribit/futures/funding/` — DeriBit Funding Data | yes | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1ohlc~1deribit~1futures~1funding~1/get) |
| `DeribitOHLC.Options(ctx, DeribitOHLCServiceOptionsParams{})` | `GET https://api.cryptodatadownload.com/v1/data/ohlc/deribit/options/` — Deribit Options Timeseries | yes | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1ohlc~1deribit~1options~1/get) |
| `DeribitOHLC.Volatility(ctx, VolatilityParams{})` | `GET https://api.cryptodatadownload.com/v1/data/ohlc/deribit/volatility/` — DeriBit Volatility Data | yes | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1ohlc~1deribit~1volatility~1/get) |

## Options Analytics

| Method | Endpoint | Raw | Docs |
|---|---|---|---|
| `OptionsAnalytics.OptionsStrikeHeatmap(ctx, OptionsStrikeHeatmapParams{})` | `GET https://api.cryptodatadownload.com/v1/data/options/strike-heatmap/` — Options Strike Heatmap (live) | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1options~1strike-heatmap~1/get) |
| `OptionsAnalytics.SummaryDeribitOptionsGreeksHigherOrder(ctx, SummaryDeribitOptionsGreeksHigherOrderParams{})` | `GET https://api.cryptodatadownload.com/v1/data/summary/deribit/options/greeks/higher-order/` — Higher-Order Greeks | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1deribit~1options~1greeks~1higher-order~1/get) |
| `OptionsAnalytics.SummaryDeribitOptionsPinCandidate(ctx, SummaryDeribitOptionsPinCandidateParams{})` | `GET https://api.cryptodatadownload.com/v1/data/summary/deribit/options/pin-candidate/` — Pin Candidate Strikes | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1deribit~1options~1pin-candidate~1/get) |
| `OptionsAnalytics.SummaryDeribitOptionsTerm(ctx, SummaryDeribitOptionsTermParams{})` | `GET https://api.cryptodatadownload.com/v1/data/summary/deribit/options/term/` — Options Term Structure | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1deribit~1options~1term~1/get) |
| `OptionsAnalytics.SummaryOptionsVwapPerStrike(ctx, SummaryOptionsVwapPerStrikeParams{})` | `GET https://api.cryptodatadownload.com/v1/data/summary/options/vwap-per-strike/` — Options VWAP per Strike-Day | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1options~1vwap-per-strike~1/get) |
| `OptionsAnalytics.VolatilityCone(ctx, VolatilityConeParams{})` | `GET https://api.cryptodatadownload.com/v1/data/volatility/cone/` — Volatility Cone | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1volatility~1cone~1/get) |
| `OptionsAnalytics.VolatilityIvRvSpread(ctx, VolatilityIvRvSpreadParams{})` | `GET https://api.cryptodatadownload.com/v1/data/volatility/iv-rv-spread/` — IV vs RV Spread | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1volatility~1iv-rv-spread~1/get) |
| `OptionsAnalytics.VolatilitySurface(ctx, VolatilitySurfaceParams{})` | `GET https://api.cryptodatadownload.com/v1/data/volatility/surface/` — Implied Volatility Surface | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1volatility~1surface~1/get) |

## Risk & Volatility Metrics

| Method | Endpoint | Raw | Docs |
|---|---|---|---|
| `Risk.RiskCorrelations(ctx, RiskCorrelationsParams{})` | `GET https://api.cryptodatadownload.com/v1/data/risk/correlations/` — Trading Correlations | yes | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1risk~1correlations~1/get) |
| `Risk.RiskCorrelationsRolling(ctx, RiskCorrelationsRollingParams{})` | `GET https://api.cryptodatadownload.com/v1/data/risk/correlations/rolling/` — Rolling Correlations | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1risk~1correlations~1rolling~1/get) |
| `Risk.RiskDrawdown(ctx, RiskDrawdownParams{})` | `GET https://api.cryptodatadownload.com/v1/data/risk/drawdown/` — Drawdown Summary | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1risk~1drawdown~1/get) |
| `Risk.RiskDrawdownUnderwater(ctx, RiskDrawdownUnderwaterParams{})` | `GET https://api.cryptodatadownload.com/v1/data/risk/drawdown/underwater/` — Underwater Curve | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1risk~1drawdown~1underwater~1/get) |
| `Risk.RiskRegime(ctx, RiskRegimeParams{})` | `GET https://api.cryptodatadownload.com/v1/data/risk/regime/` — Market Regime Label | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1risk~1regime~1/get) |
| `Risk.RiskVarByMethod(ctx, RiskVarByMethodParams{})` | `GET https://api.cryptodatadownload.com/v1/data/risk/var/by-method/` — Value at Risk (VaR) | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1risk~1var~1by-method~1/get) |
| `Risk.RiskVarPortfolio(ctx, RiskVarPortfolioParams{})` | `GET https://api.cryptodatadownload.com/v1/data/risk/var/portfolio/` — Portfolio Value at Risk | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1risk~1var~1portfolio~1/get) |
| `Risk.RiskVarStandalone(ctx, RiskVarStandaloneParams{})` | `GET https://api.cryptodatadownload.com/v1/data/risk/var/standalone/` — Value at Risk by Symbol | yes | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1risk~1var~1standalone~1/get) |
| `Risk.VolatilityRealized(ctx, VolatilityRealizedParams{})` | `GET https://api.cryptodatadownload.com/v1/data/volatility/realized/` — Realized Volatility | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1volatility~1realized~1/get) |

## Binance Summary

| Method | Endpoint | Raw | Docs |
|---|---|---|---|
| `BinanceSummary.Basis(ctx, BasisParams{})` | `GET https://api.cryptodatadownload.com/v1/data/summary/binance/basis/` — Spot and Futures Prices & Corresponding Basis | yes | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1binance~1basis~1/get) |
| `BinanceSummary.FuturesMetrics(ctx, FuturesMetricsParams{})` | `GET https://api.cryptodatadownload.com/v1/data/summary/binance/futures/metrics/` — Binance Futures UM Summary Metrics | yes | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1binance~1futures~1metrics~1/get) |
| `BinanceSummary.OptionsOi(ctx, OptionsOiParams{})` | `GET https://api.cryptodatadownload.com/v1/data/summary/binance/options/oi/` — Binance Options OI Totals | yes | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1binance~1options~1oi~1/get) |
| `BinanceSummary.OptionsOiInstrument(ctx, OptionsOiInstrumentParams{})` | `GET https://api.cryptodatadownload.com/v1/data/summary/binance/options/oi/instrument/` — Binance Options OI Totals by Instrument | yes | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1binance~1options~1oi~1instrument~1/get) |
| `BinanceSummary.OptionsOiMaturity(ctx, OptionsOiMaturityParams{})` | `GET https://api.cryptodatadownload.com/v1/data/summary/binance/options/oi/maturity/` — Binance Options OI Totals by Maturity | yes | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1binance~1options~1oi~1maturity~1/get) |
| `BinanceSummary.SpotFlow(ctx, SpotFlowParams{})` | `GET https://api.cryptodatadownload.com/v1/data/summary/binance/spot/flow/` — Binance Flow Imbalance (Order Flow Delta) | yes | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1binance~1spot~1flow~1/get) |
| `BinanceSummary.SpotLargeTrades(ctx, SpotLargeTradesParams{})` | `GET https://api.cryptodatadownload.com/v1/data/summary/binance/spot/large-trades/` — Binance Large Trades (Whale Activity) | yes | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1binance~1spot~1large-trades~1/get) |
| `BinanceSummary.SpotTransactional(ctx, SpotTransactionalParams{})` | `GET https://api.cryptodatadownload.com/v1/data/summary/binance/spot/transactional/` — Binance Spot Volume Summary | yes | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1binance~1spot~1transactional~1/get) |
| `BinanceSummary.SpotTransactionalDaily(ctx, SpotTransactionalDailyParams{})` | `GET https://api.cryptodatadownload.com/v1/data/summary/binance/spot/transactional/daily/` — Binance Daily Trade Summary | yes | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1binance~1spot~1transactional~1daily~1/get) |
| `BinanceSummary.SpotTransactionalHourly(ctx, SpotTransactionalHourlyParams{})` | `GET https://api.cryptodatadownload.com/v1/data/summary/binance/spot/transactional/hourly/` — Binance Hourly Trade Summary | yes | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1binance~1spot~1transactional~1hourly~1/get) |
| `BinanceSummary.SpotVolumeProfile(ctx, SpotVolumeProfileParams{})` | `GET https://api.cryptodatadownload.com/v1/data/summary/binance/spot/volume-profile/` — Binance Volume Profile (Price Distribution) | yes | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1binance~1spot~1volume-profile~1/get) |

## DeriBit Summary

| Method | Endpoint | Raw | Docs |
|---|---|---|---|
| `DeribitSummary.SummaryDeribitOptionsFlow(ctx, SummaryDeribitOptionsFlowParams{})` | `GET https://api.cryptodatadownload.com/v1/data/summary/deribit/options/flow/` — Deribit Options Flow Analysis | yes | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1deribit~1options~1flow~1/get) |
| `DeribitSummary.SummaryDeribitOptionsGreeks(ctx, SummaryDeribitOptionsGreeksParams{})` | `GET https://api.cryptodatadownload.com/v1/data/summary/deribit/options/greeks/` — DeriBit Greeks Summary by Instrument | yes | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1deribit~1options~1greeks~1/get) |
| `DeribitSummary.SummaryDeribitOptionsGreeksMaturities(ctx, SummaryDeribitOptionsGreeksMaturitiesParams{})` | `GET https://api.cryptodatadownload.com/v1/data/summary/deribit/options/greeks/maturities/` — DeriBit Greeks Summary by Maturity | yes | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1deribit~1options~1greeks~1maturities~1/get) |
| `DeribitSummary.SummaryDeribitOptionsLargeTrades(ctx, SummaryDeribitOptionsLargeTradesParams{})` | `GET https://api.cryptodatadownload.com/v1/data/summary/deribit/options/large-trades/` — Deribit Options Large Trades Analysis | yes | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1deribit~1options~1large-trades~1/get) |
| `DeribitSummary.SummaryDeribitOptionsRatios(ctx, SummaryDeribitOptionsRatiosParams{})` | `GET https://api.cryptodatadownload.com/v1/data/summary/deribit/options/ratios/` — Deribit Options Put/Call Ratios | yes | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1deribit~1options~1ratios~1/get) |
| `DeribitSummary.SummaryDeribitOptionsTransactionsLargest(ctx, SummaryDeribitOptionsTransactionsLargestParams{})` | `GET https://api.cryptodatadownload.com/v1/data/summary/deribit/options/transactions/largest/` — Top 100 DeriBit Option Transactions per Day in USD | yes | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1deribit~1options~1transactions~1largest~1/get) |
| `DeribitSummary.SummaryDeribitOptionsVolume(ctx, SummaryDeribitOptionsVolumeParams{})` | `GET https://api.cryptodatadownload.com/v1/data/summary/deribit/options/volume/` — Deribit Options Volume Summary | yes | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1summary~1deribit~1options~1volume~1/get) |
| `DeribitSummary.TransactionsDeribitOptions(ctx, TransactionsDeribitOptionsParams{})` | `GET https://api.cryptodatadownload.com/v1/data/transactions/deribit/options/` — Deribit Options Transactions | yes | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1transactions~1deribit~1options~1/get) |

## Technical Indicators

| Method | Endpoint | Raw | Docs |
|---|---|---|---|
| `Technicals.TechnicalsTA(ctx, TechnicalsTAParams{})` | `GET https://api.cryptodatadownload.com/v1/data/technicals/TA/` — Technical Analysis Indicators | yes | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1data~1technicals~1TA~1/get) |

## Options Pricing Tools

| Method | Endpoint | Raw | Docs |
|---|---|---|---|
| `OptionsPricing.PricingGreeks(ctx, PricingGreeksParams{})` | `GET https://api.cryptodatadownload.com/v1/risk/options/pricing/greeks/` — Black Scholes Merton (BSM) Option Pricing Model | — | [UI](https://api.cryptodatadownload.com/) · [schema](https://api.cryptodatadownload.com/?format=openapi#/paths/~1risk~1options~1pricing~1greeks~1/get) |
