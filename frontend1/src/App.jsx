import React, { useEffect, useState } from "react";
import axios from "axios";
import {
    ResponsiveContainer,
    ComposedChart,
    Line,
    Area,
    Bar,
    XAxis,
    YAxis,
    CartesianGrid,
    Tooltip,
    Legend,
} from "recharts";

export default function App() {
    // Store the stock data from the backend
    const [chartData, setChartData] = useState([]);
    // Which stock we want to look at (you can change this later to a dropdown)
    const [stockSymbol, setStockSymbol] = useState("AAPL");

    // Load data from backend when the component loads OR when stockSymbol changes
    useEffect(() => {
        async function fetchStockData() {
            try {
                const response = await axios.get(
                    `http://localhost:8080/api/v1/chart/${stockSymbol}`
                );

                // Convert backend data into a simpler shape for charts
                const preparedData = response.data.data.map((point) => ({
                    date: new Date(point.bar.timestamp).toLocaleDateString(),
                    open: point.bar.open,
                    high: point.bar.high,
                    low: point.bar.low,
                    close: point.bar.close,
                    volume: point.bar.volume,

                    // Indicators (sometimes missing, so we use ? to avoid crashes)
                    rsi: point.indicators?.rsi,
                    sma20: point.indicators?.sma_20,
                    ema20: point.indicators?.ema_20,
                    upperBB: point.indicators?.bollinger_bands?.upper,
                    middleBB: point.indicators?.bollinger_bands?.middle,
                    lowerBB: point.indicators?.bollinger_bands?.lower,
                    macd: point.indicators?.macd?.macd,
                    macdSignal: point.indicators?.macd?.signal,
                    macdHist: point.indicators?.macd?.histogram,
                }));

                setChartData(preparedData);
            } catch (error) {
                console.error("Could not load stock data:", error);
            }
        }

        fetchStockData();
    }, [stockSymbol]);

    return (
        <div className="p-6 font-sans">
            {/* Title */}
            <h1 className="text-2xl font-bold mb-6">
                Stock Dashboard – {stockSymbol}
            </h1>

            {/* MAIN PRICE CHART */}
            <div className="mb-10">
                <h2 className="text-lg font-semibold mb-2">Price + Indicators</h2>
                <ResponsiveContainer width="100%" height={400}>
                    <ComposedChart data={chartData}>
                        <CartesianGrid strokeDasharray="3 3" />
                        <XAxis dataKey="date" minTickGap={20} />
                        <YAxis />
                        <Tooltip />
                        <Legend />

                        {/* Price Line (close price over time) */}
                        <Area
                            type="monotone"
                            dataKey="close"
                            stroke="#000"
                            fill="#8884d8"
                            fillOpacity={0.2}
                            name="Close"
                        />

                        {/* Bollinger Bands */}
                        <Line dataKey="upperBB" stroke="red" dot={false} name="Upper BB" />
                        <Line
                            dataKey="middleBB"
                            stroke="orange"
                            dot={false}
                            name="Middle BB"
                        />
                        <Line dataKey="lowerBB" stroke="green" dot={false} name="Lower BB" />

                        {/* SMA & EMA */}
                        <Line dataKey="sma20" stroke="blue" dot={false} name="SMA 20" />
                        <Line dataKey="ema20" stroke="purple" dot={false} name="EMA 20" />
                    </ComposedChart>
                </ResponsiveContainer>
            </div>

            {/* RSI CHART */}
            <div className="mb-10">
                <h2 className="text-lg font-semibold mb-2">RSI (Relative Strength Index)</h2>
                <ResponsiveContainer width="100%" height={200}>
                    <ComposedChart data={chartData}>
                        <CartesianGrid strokeDasharray="3 3" />
                        <XAxis dataKey="date" minTickGap={20} />
                        <YAxis domain={[0, 100]} />
                        <Tooltip />
                        <Line dataKey="rsi" stroke="purple" dot={false} name="RSI" />
                    </ComposedChart>
                </ResponsiveContainer>
            </div>

            {/* MACD CHART */}
            <div>
                <h2 className="text-lg font-semibold mb-2">MACD (Moving Average Convergence Divergence)</h2>
                <ResponsiveContainer width="100%" height={250}>
                    <ComposedChart data={chartData}>
                        <CartesianGrid strokeDasharray="3 3" />
                        <XAxis dataKey="date" minTickGap={20} />
                        <YAxis />
                        <Tooltip />
                        <Legend />
                        <Line dataKey="macd" stroke="blue" dot={false} name="MACD" />
                        <Line
                            dataKey="macdSignal"
                            stroke="red"
                            dot={false}
                            name="Signal"
                        />
                        <Bar dataKey="macdHist" fill="gray" name="Histogram" />
                    </ComposedChart>
                </ResponsiveContainer>
            </div>
        </div>
    );
}
