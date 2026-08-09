import axios from "axios"

const configuredApiURL = process.env.NEXT_PUBLIC_API_URL ?? ""
export const apiURL = configuredApiURL.replace(/\/$/, "")

export const apiClient = axios.create({
  baseURL: apiURL,
  headers: {
    "Content-Type": "application/json",
  },
})

apiClient.interceptors.request.use(
  (config) => {
    if (typeof window !== "undefined") {
      const token =
        window.localStorage.getItem("monitoring_token") ??
        window.localStorage.getItem("token")
      if (token && config.headers) {
        config.headers.Authorization = `Bearer ${token}`
      }
    }
    return config;
  },
  (error) => Promise.reject(error)
)

apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    const message =
      error.response?.data?.error ||
      error.message ||
      "An unexpected error occurred."
    return Promise.reject(new Error(message))
  }
)

export default apiClient
