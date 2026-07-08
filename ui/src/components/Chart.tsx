import { useTheme } from '@mui/material/styles';
import { AreaChart, XAxis, YAxis, Label, ResponsiveContainer, Area, CartesianGrid } from 'recharts';
import { Title } from './Title.tsx';
import { WalletChartBalance } from '../utils/model.ts';
import { Fragment } from 'react';
import { useMediaQuery } from '@mui/material';

interface ChartProps {
  data: WalletChartBalance[]
}

export function Chart(props: ChartProps) {

  const theme = useTheme();
  const isXs = useMediaQuery(theme.breakpoints.only('xs'));
  const maxPoints = isXs ? 3 : 6;
  const chartData = [...props.data].sort((a, b) => a.date - b.date).slice(-maxPoints);
  const maxSum = Math.max(0, ...chartData.map((item) => item.sum));
  const roundedMax = Math.max(5000, Math.ceil(maxSum / 5000) * 5000);
  const yTicks = Array.from({ length: roundedMax / 5000 + 1 }, (_, index) => index * 5000);

  return (
    <Fragment>
      <Title>Balance</Title>
      <ResponsiveContainer>
        <AreaChart
          data={chartData}
          margin={{
            top: 16,
            right: 16,
            bottom: 0,
            left: 24,
          }}
        >
          <XAxis
            dataKey="date"
            stroke={theme.palette.text.secondary}
            style={theme.typography.body2}
          />
          <YAxis
            dataKey="sum"
            stroke={theme.palette.text.secondary}
            style={theme.typography.body2}
            domain={[0, roundedMax]}
            ticks={yTicks}
            tickFormatter={(value) => (value === 0 ? '0' : `${value / 1000}k`)}
          >
            <Label
              angle={270}
              position="left"
              style={{
                textAnchor: 'middle',
                fill: theme.palette.text.primary,
                ...theme.typography.body1,
              }}
            >
              Balance (S$)
            </Label>
          </YAxis>
          <CartesianGrid strokeDasharray="3 3" />
          <Area
            isAnimationActive={false}
            type="monotone"
            dataKey="sum"
            stroke={theme.palette.primary.main}
            dot={true}
          />
        </AreaChart>
      </ResponsiveContainer>
    </Fragment>
  );
}