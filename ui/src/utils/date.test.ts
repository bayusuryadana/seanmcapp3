import {
  currentYearMonth,
  formatYearMonth,
  isValidYearMonth,
  shiftYearMonth,
  yearMonthTitle,
} from './date'

describe('date utils', () => {
  it('formatYearMonth pads the month', () => {
    expect(formatYearMonth(new Date(2024, 5, 15))).toBe('202406') // June
    expect(formatYearMonth(new Date(2024, 11, 1))).toBe('202412') // December
  })

  it('currentYearMonth matches formatYearMonth(now)', () => {
    expect(currentYearMonth()).toBe(formatYearMonth(new Date()))
  })

  it('shiftYearMonth handles year boundaries', () => {
    expect(shiftYearMonth('202406', 1)).toBe('202407')
    expect(shiftYearMonth('202406', -1)).toBe('202405')
    expect(shiftYearMonth('202412', 1)).toBe('202501')
    expect(shiftYearMonth('202401', -1)).toBe('202312')
  })

  it('yearMonthTitle renders month name + year', () => {
    expect(yearMonthTitle('202406')).toBe('June 2024')
    expect(yearMonthTitle('202401')).toBe('January 2024')
  })

  it('isValidYearMonth checks for exactly 6 digits', () => {
    expect(isValidYearMonth('202406')).toBe(true)
    expect(isValidYearMonth('20246')).toBe(false)
    expect(isValidYearMonth('2024066')).toBe(false)
    expect(isValidYearMonth('abcdef')).toBe(false)
  })
})

