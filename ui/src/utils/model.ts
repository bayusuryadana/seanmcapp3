export type WalletDashboardData = {
  chart: WalletChart;
  allocations: WalletAllocations[];
  savings: WalletSavings;
  planned: WalletPlanned;
  detail: WalletDetail[];
}

export type WalletChart = {
  balance: WalletChartBalance[];
}

export type WalletAllocations = {
  name: string,
  expense: number,
  alloc: number,
}

export type WalletChartBalance = {
  date: number;
  sum: number;
}

export type WalletSavings = {
  dbs: number;
  bca: number;
}

export type WalletPlanned = {
  sgd: number;
  idr: number;
}

export type WalletDetail = {
  id: number;
  date: number;
  name: string;
  category: string;
  currency: string;
  amount: number;
  done: boolean;
  account: string;
}

export type WalletAlert = {
  display: string
  text: string
}