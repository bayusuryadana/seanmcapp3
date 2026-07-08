#!/usr/bin/env node
// Computes Go statement coverage from a coverprofile, EXCLUDING pure
// wiring/config/entrypoint files (dependency injection, env loading, router &
// scheduler bootstrapping, main). Fails if the result is below COVERAGE_THRESHOLD.

import { readFileSync } from 'node:fs'

const THRESHOLD = Number(process.env.COVERAGE_THRESHOLD ?? 90)
const profile = process.env.GO_COVERAGE ?? 'coverage.out'

// Files that are glue/injection/config and not meaningful to unit test.
const EXCLUDE = new Set([
  'seanmcapp/main.go',
  'seanmcapp/bootstrap/injection.go',
  'seanmcapp/bootstrap/server.go',
  'seanmcapp/util/settings.go',
])

const text = readFileSync(profile, 'utf8')
let covered = 0
let total = 0
for (const line of text.split('\n')) {
  if (!line || line.startsWith('mode:')) continue
  const file = line.slice(0, line.indexOf(':'))
  if (EXCLUDE.has(file)) continue
  const parts = line.trim().split(' ')
  const count = Number(parts[parts.length - 1])
  const stmts = Number(parts[parts.length - 2])
  if (!Number.isFinite(count) || !Number.isFinite(stmts)) continue
  total += stmts
  if (count > 0) covered += stmts
}

const pct = total === 0 ? 100 : (covered / total) * 100
console.log(`Backend coverage (excluding wiring/config): ${covered}/${total}  ${pct.toFixed(2)}%`)
console.log(`Threshold: ${THRESHOLD}%`)

if (pct < THRESHOLD) {
  console.error(`\n❌ Backend coverage ${pct.toFixed(2)}% is below the ${THRESHOLD}% threshold`)
  process.exit(1)
}
console.log(`\n✅ Backend coverage ${pct.toFixed(2)}% meets the ${THRESHOLD}% threshold`)

