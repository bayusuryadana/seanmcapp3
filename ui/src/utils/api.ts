import axios from "axios"
import { API_URL } from "./constant"

// Single preconfigured axios instance used across the app.
export const api = axios.create({ baseURL: API_URL })

// Allow the app to register what should happen on a 401 (e.g. log the user out).
let onUnauthorized: (() => void) | null = null
export const setUnauthorizedHandler = (handler: (() => void) | null) => {
  onUnauthorized = handler
}

// Attach the auth token (if any) to every request.
api.interceptors.request.use((config) => {
  const token = localStorage.getItem("token")
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// Handle expired/invalid sessions in a single place.
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (axios.isAxiosError(error) && error.response?.status === 401) {
      onUnauthorized?.()
    }
    return Promise.reject(error)
  }
)

