import ReactDOM from 'react-dom/client'
import { Home } from './Home'
import './index.css'
import { createBrowserRouter, RouterProvider } from "react-router-dom"
import { Wallet } from './pages/Wallet'
import { WalletLogin } from './pages/WalletLogin'
import { UserProvider } from './UserContext'
import { StrictMode } from 'react'
import { WalletDashboard } from './pages/WalletDashboard'

const router = createBrowserRouter([
  {
    path: "/",
    element: <Home />
  },
  {
    path: "/wallet/login",
    element: <WalletLogin />
  },
  {
    element: <Wallet />,
    children: [
      {
        path: "/wallet",
        element: <WalletDashboard />
      },
    ]
  },
]);

ReactDOM.createRoot(document.getElementById('root')!).render(
  <StrictMode>
      <UserProvider>
        <RouterProvider router={router} />
      </UserProvider>
  </StrictMode>,
)
