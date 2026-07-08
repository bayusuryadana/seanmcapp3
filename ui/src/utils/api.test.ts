import axios, { AxiosResponse, InternalAxiosRequestConfig } from 'axios'
import { api, setUnauthorizedHandler } from './api'

describe('api instance', () => {
  afterEach(() => {
    localStorage.clear()
    setUnauthorizedHandler(null)
  })

  it('attaches the bearer token from localStorage', async () => {
    localStorage.setItem('token', 'abc123')
    let seen: InternalAxiosRequestConfig | undefined

    api.defaults.adapter = async (config) => {
      seen = config as InternalAxiosRequestConfig
      return { data: {}, status: 200, statusText: 'OK', headers: {}, config } as AxiosResponse
    }

    await api.get('/anything')
    expect(seen?.headers.Authorization).toBe('Bearer abc123')
  })

  it('omits Authorization when no token is stored', async () => {
    let seen: InternalAxiosRequestConfig | undefined
    api.defaults.adapter = async (config) => {
      seen = config as InternalAxiosRequestConfig
      return { data: {}, status: 200, statusText: 'OK', headers: {}, config } as AxiosResponse
    }

    await api.get('/anything')
    expect(seen?.headers.Authorization).toBeUndefined()
  })

  it('calls the unauthorized handler on a 401 response', async () => {
    const onUnauthorized = jest.fn()
    setUnauthorizedHandler(onUnauthorized)

    api.defaults.adapter = async (config) => {
      throw new axios.AxiosError('unauth', 'ERR_BAD_REQUEST', config, null, {
        status: 401,
        data: {},
        statusText: 'Unauthorized',
        headers: {},
        config,
      } as AxiosResponse)
    }

    await expect(api.get('/secure')).rejects.toBeDefined()
    expect(onUnauthorized).toHaveBeenCalledTimes(1)
  })

  it('does not call the handler on non-401 errors', async () => {
    const onUnauthorized = jest.fn()
    setUnauthorizedHandler(onUnauthorized)

    api.defaults.adapter = async (config) => {
      throw new axios.AxiosError('boom', 'ERR', config, null, {
        status: 500,
        data: {},
        statusText: 'Server Error',
        headers: {},
        config,
      } as AxiosResponse)
    }

    await expect(api.get('/secure')).rejects.toBeDefined()
    expect(onUnauthorized).not.toHaveBeenCalled()
  })
})

