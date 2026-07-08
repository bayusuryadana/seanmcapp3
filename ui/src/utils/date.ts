// Helpers for the `YYYYMM` string format used by the wallet dashboard.

const MONTH_NAMES = [
  "January", "February", "March", "April", "May", "June",
  "July", "August", "September", "October", "November", "December",
]

export const formatYearMonth = (date: Date): string =>
  `${date.getFullYear()}${String(date.getMonth() + 1).padStart(2, "0")}`

export const currentYearMonth = (): string => formatYearMonth(new Date())

// Shift a `YYYYMM` value by a number of months (handles year boundaries).
export const shiftYearMonth = (yearMonth: string, delta: number): string => {
  const year = Number(yearMonth.slice(0, 4))
  const month = Number(yearMonth.slice(4, 6))
  return formatYearMonth(new Date(year, month - 1 + delta, 1))
}

export const yearMonthTitle = (yearMonth: string): string => {
  const monthName = MONTH_NAMES[Number(yearMonth.slice(4, 6)) - 1] ?? ""
  return `${monthName} ${yearMonth.slice(0, 4)}`
}

export const isValidYearMonth = (value: string): boolean => /^\d{6}$/.test(value)

