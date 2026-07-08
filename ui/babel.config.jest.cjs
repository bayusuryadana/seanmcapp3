// Babel config used ONLY by Jest (referenced explicitly from jest.config.cjs).
// Vite keeps using @vitejs/plugin-react and never loads this file.

// Rewrites Vite's `import.meta.env` -> `process.env` so Jest can evaluate modules
// (e.g. utils/constant.ts) that read Vite env vars.
function importMetaEnvPlugin() {
  return {
    visitor: {
      MetaProperty(path) {
        path.replaceWithSourceString('({ env: process.env })')
      },
    },
  }
}

module.exports = {
  presets: [
    ['@babel/preset-env', { targets: { node: 'current' } }],
    ['@babel/preset-react', { runtime: 'automatic' }],
    '@babel/preset-typescript',
  ],
  plugins: [importMetaEnvPlugin],
}

