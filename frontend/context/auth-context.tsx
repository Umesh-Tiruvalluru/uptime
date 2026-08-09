"use client"

import React, { createContext, useContext, useEffect, useState, useCallback } from "react"
import {
  getMe,
  loginUser,
  registerUser,
  type User,
  type AuthResponse,
} from "@/lib/monitoring"

interface AuthContextType {
  user: User | null
  token: string | null
  isLoading: boolean
  login: (email: string, password: string) => Promise<void>
  register: (
    firstName: string,
    lastName: string,
    email: string,
    password: string
  ) => Promise<void>
  logout: () => void
  refreshUser: () => Promise<void>
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export const AUTH_TOKEN_KEY = "monitoring_token"

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [token, setToken] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState<boolean>(true)

  const loadUser = useCallback(async (authToken: string) => {
    try {
      const userData = await getMe()
      setUser(userData)
    } catch {
      // Token invalid or expired
      window.localStorage.removeItem(AUTH_TOKEN_KEY)
      window.localStorage.removeItem("token")
      setToken(null)
      setUser(null)
    }
  }, [])

  useEffect(() => {
    const savedToken =
      window.localStorage.getItem(AUTH_TOKEN_KEY) ??
      window.localStorage.getItem("token")

    if (savedToken) {
      setToken(savedToken)
      void loadUser(savedToken).finally(() => setIsLoading(false))
    } else {
      setIsLoading(false)
    }
  }, [loadUser])

  const login = async (email: string, password: string) => {
    const res = await loginUser({ email, password })
    if (res.token) {
      window.localStorage.setItem(AUTH_TOKEN_KEY, res.token)
      setToken(res.token)
      await loadUser(res.token)
    }
  }

  const register = async (
    firstName: string,
    lastName: string,
    email: string,
    password: string
  ) => {
    const res: AuthResponse = await registerUser({
      firstName,
      lastName,
      email,
      password,
    })
    if (res.token) {
      window.localStorage.setItem(AUTH_TOKEN_KEY, res.token)
      setToken(res.token)
      if (res.user) {
        setUser(res.user)
      } else {
        await loadUser(res.token)
      }
    }
  }

  const logout = () => {
    window.localStorage.removeItem(AUTH_TOKEN_KEY)
    window.localStorage.removeItem("token")
    setToken(null)
    setUser(null)
  }

  const refreshUser = async () => {
    if (token) {
      await loadUser(token)
    }
  }

  return (
    <AuthContext.Provider
      value={{
        user,
        token,
        isLoading,
        login,
        register,
        logout,
        refreshUser,
      }}
    >
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error("useAuth must be used within an AuthProvider")
  }
  return context
}
